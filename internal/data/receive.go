package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/omalloc/trust-receive/internal/biz"
)

var _ biz.ReceiveRepo = (*receiveRepo)(nil)

type receiveRepo struct {
	data *Data
	log  *log.Helper
}

func NewReceiveRepo(data *Data, logger log.Logger) biz.ReceiveRepo {
	return &receiveRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Save persists file information to Redis
// Storage format: key = file:{hash_key}, value = {hash}
func (r *receiveRepo) Save(ctx context.Context, key string, hash string) error {
	return r.data.rdb.Set(ctx, "file:"+key, hash, 0).Err()
}

// Get retrieves stored file HASH from Redis
func (r *receiveRepo) Get(ctx context.Context, key string) (string, error) {
	hash, err := r.data.rdb.Get(ctx, "file:"+key).Result()
	if err != nil {
		return "", err
	}
	return hash, nil
}
