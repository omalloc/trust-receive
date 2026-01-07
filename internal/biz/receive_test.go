package biz

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

type mockRepo struct {
	data map[string]string // { key:value }
}

func (m *mockRepo) Save(_ context.Context, key string, hash string) error {
	m.data[key] = hash
	return nil
}

func (m *mockRepo) Get(_ context.Context, key string) (string, error) {
	if hash, ok := m.data[key]; ok {
		return hash, nil
	}
	return "", fmt.Errorf("not found")
}

type mockAlerter struct {
	alerted bool
}

func (m *mockAlerter) Alert(ctx context.Context, msg string) error {
	m.alerted = true
	return nil
}

func TestReceiveUsecase_Verify(t *testing.T) {
	repo := &mockRepo{data: make(map[string]string)}
	alerter := &mockAlerter{}
	uc := NewReceiveUsecase(repo, alerter, log.DefaultLogger)

	ctx := context.Background()

	// 1. First report
	info1 := &FileInfo{URL: "https://example.com/a.txt", Hash: "hash1", FileSize: 100, LastModified: "2023-01-01"}
	err := uc.Verify(ctx, info1)
	if err != nil {
		t.Fatalf("first report failed: %v", err)
	}
	key1 := hashKey(info1)
	if repo.data[key1] == "" {
		t.Fatal("data not saved in repo")
	}

	// 2. Same URL + LM report again, consistent
	err = uc.Verify(ctx, info1)
	if err != nil {
		t.Fatalf("consistent report failed: %v", err)
	}
	if alerter.alerted {
		t.Fatal("should not alert for consistent data")
	}

	// 3. Same URL + LM report again, Hash inconsistent
	info2 := &FileInfo{URL: "https://example.com/a.txt", Hash: "hash2", FileSize: 100, LastModified: "2023-01-01"}
	err = uc.Verify(ctx, info2)
	if err == nil {
		t.Fatal("should return error for inconsistent hash")
	}
	if !alerter.alerted {
		t.Fatal("should alert for inconsistent hash")
	}

	// 4. Different LastModified should be treated as a new record (no error)
	alerter.alerted = false
	info3 := &FileInfo{URL: "https://example.com/a.txt", Hash: "hash2", FileSize: 100, LastModified: "2023-01-02"}
	err = uc.Verify(ctx, info3)
	if err != nil {
		t.Fatalf("report with different LM failed: %v", err)
	}
	if alerter.alerted {
		t.Fatal("should not alert for different LM")
	}
	key3 := hashKey(info3)
	if repo.data[key3] == "" {
		t.Fatal("data for different LM not saved")
	}
	if key1 == key3 {
		t.Fatal("keys for different LM should be different")
	}

	// 5. Reset alert status, test Size inconsistency
	alerter.alerted = false
	info4 := &FileInfo{URL: "https://example.com/b.txt", Hash: "hash1", FileSize: 100, LastModified: "2023-01-01"}
	_ = uc.Verify(ctx, info4) // Store b.txt

	info5 := &FileInfo{URL: "https://example.com/b.txt", Hash: "hash1", FileSize: 200, LastModified: "2023-01-01"}
	err = uc.Verify(ctx, info5)
	if err == nil {
		t.Fatal("should return error for inconsistent size")
	}
	if !alerter.alerted {
		t.Fatal("should alert for inconsistent size")
	}
}
