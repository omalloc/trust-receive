package data

import (
	"github.com/omalloc/trust-receive/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewReceiveRepo,
)

// Data .
type Data struct {
	rdb redis.UniversalClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	log := log.NewHelper(logger)

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        c.Redis.Addrs,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})
	cleanup := func() {
		log.Info("closing the data resources")
		if err := rdb.Close(); err != nil {
			log.Error(err)
		}
	}
	return &Data{rdb: rdb}, cleanup, nil
}
