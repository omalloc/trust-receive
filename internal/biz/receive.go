package biz

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_metricVerifiedConflictTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tr_",
		Subsystem: "trust_receive",
		Name:      "verified_conflict_total",
		Help:      "Total number of file verification conflicts detected",
	}, []string{"url", "hash", "conflict_hash", "last_modified", "file_size"})
)

// FileInfo stores file verification information
type FileInfo struct {
	URL          string
	Hash         string
	FileSize     uint64
	LastModified string
}

// ReceiveRepo defines the data persistence interface
type ReceiveRepo interface {
	// Save stores a hash string associated with a unique key in the data store.
	// Returns an error if the operation fails.
	Save(ctx context.Context, key string, hash string) error
	// Get retrieves the hash string associated with a unique key from the data store.
	// Returns an error if the operation fails.
	Get(ctx context.Context, key string) (string, error)
}

// Alerter defines the alert handling interface
type Alerter interface {
	Alert(ctx context.Context, msg string) error
}

func init() {
	prometheus.MustRegister(_metricVerifiedConflictTotal)
}

// ReceiveUsecase is the business usecase for handling reception logic
type ReceiveUsecase struct {
	repo    ReceiveRepo
	alerter Alerter
	bf      *bloom.BloomFilter // Bloom Filter for fast deduplication, reducing Redis query pressure
	log     *log.Helper
}

// NewReceiveUsecase creates a new ReceiveUsecase
func NewReceiveUsecase(repo ReceiveRepo, alerter Alerter, logger log.Logger) *ReceiveUsecase {
	return &ReceiveUsecase{
		repo:    repo,
		alerter: alerter,
		bf:      bloom.NewWithEstimates(1000000, 0.01), // Initial capacity 1M, error rate 1%
		log:     log.NewHelper(logger),
	}
}

// Verify performs real-time file consistency checking
// Logical flow:
// 1. Generate a unique Key based on URL + LastModified.
// 2. Use Bloom Filter to quickly determine if it's the first time seeing this Key.
// 3. If it's the first time, save directly to Redis persistent storage.
// 4. If Bloom Filter suggests it exists, query Redis for actual data and compare.
// 5. If stored data differs from reported data (Hash + Size), trigger an alert.
func (uc *ReceiveUsecase) Verify(ctx context.Context, info *FileInfo) error {
	key := hashKey(info)
	hash := hashValue(info)

	// 1. Use Bloom Filter to quickly check if it might have been reported
	// TestOrAdd returns true if the element already exists
	if !uc.bf.TestOrAdd([]byte(key)) {
		// If it doesn't exist, it's the first time seeing this combination, save to Redis
		uc.log.Infof("New file version reported: URL=%s, LM=%s, hash: %s, size: %d", info.URL, info.LastModified, hash, info.FileSize)
		return uc.repo.Save(ctx, key, hash)
	}

	// 2. Bloom Filter suggests it might exist, query actual data from Redis
	storedHash, err := uc.repo.Get(ctx, key)
	if err != nil {
		// If no record in Redis (Bloom Filter false positive), save as a new record
		uc.log.Infof("Key %s not in store (Bloom Filter false positive), saving", key)
		return uc.repo.Save(ctx, key, hash)
	}

	// 3. Detect inconsistency: same combination with different hash or fileSize
	if storedHash != hash {
		msg := fmt.Sprintf("CRITICAL: File inconsistency detected for URL: %s (LM: %s) Stored: [Hash: %s] Reported: [Hash: %s]",
			info.URL, info.LastModified, storedHash, hash)

		_metricVerifiedConflictTotal.WithLabelValues(info.URL, storedHash, hash, info.LastModified, fmt.Sprintf("%d", info.FileSize)).Inc()

		// Trigger alert or execute command
		if uc.alerter != nil {
			_ = uc.alerter.Alert(ctx, msg)
		}
		return fmt.Errorf("file inconsistency")
	}

	uc.log.Debugf("URL %s (LM: %s) consistency check passed", info.URL, info.LastModified)
	return nil
}

// hashKey generates a unique identifier based on URL and Last-Modified to detect the same file version
func hashKey(info *FileInfo) string {
	h := sha256.New()
	h.Write([]byte(info.URL))
	h.Write([]byte(info.LastModified))
	return hex.EncodeToString(h.Sum(nil))
}

func hashValue(info *FileInfo) string {
	h := md5.New()
	h.Write([]byte(info.Hash))
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], info.FileSize)
	h.Write(buf[:])
	return hex.EncodeToString(h.Sum(nil))
}
