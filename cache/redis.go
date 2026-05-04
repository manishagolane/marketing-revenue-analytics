package cache

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/rueidis"
	"go.uber.org/zap"

	"marketing-revenue-analytics/config"
	"marketing-revenue-analytics/models"
)

type Cache struct {
	rDB     rueidis.Client
	queries *models.Queries
	log     *zap.Logger
}

var ErrorNotFound = errors.New("not found")

func NewCache(queries *models.Queries, log *zap.Logger) *Cache {
	client, err := newRedisClient()
	if err != nil {
		log.Fatal("error creating redis client", zap.Error(err))
	}
	return &Cache{rDB: client, queries: queries, log: log}
}

func newRedisClient() (rueidis.Client, error) {
	return rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{config.GetString("redis.address")},
		SelectDB:     config.GetInt("redis.db"),
		Password:     config.GetString("redis.password"),
		AlwaysRESP2:  true,
		DisableCache: true,
	})
}

func (c *Cache) GetClient() rueidis.Client { return c.rDB }

func (c *Cache) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.rDB.Do(ctx, c.buildSetCmd(key, value, expiration)).Error()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rDB.Do(ctx, c.rDB.B().Get().Key(key).Build()).ToString()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return "", ErrorNotFound
		}
		return "", err
	}
	return val, nil
}

func (c *Cache) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	v := "0"
	if value {
		v = "1"
	}
	return c.rDB.Do(ctx, c.buildSetCmd(key, v, expiration)).Error()
}

func (c *Cache) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rDB.Do(ctx, c.buildSetCmd(key, string(b), expiration)).Error()
}

func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	reader, err := c.rDB.Do(ctx, c.rDB.B().Get().Key(key).Build()).AsReader()
	if err != nil {
		if rueidis.IsRedisNil(err) {
			return ErrorNotFound
		}
		return err
	}
	return json.NewDecoder(reader).Decode(dest)
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.rDB.Do(ctx, c.rDB.B().Del().Key(key).Build()).Error()
}

func (c *Cache) Close() { c.rDB.Close() }

func (c *Cache) buildSetCmd(key, value string, expiration time.Duration) rueidis.Completed {
	if expiration <= 0 {
		return c.rDB.B().Set().Key(key).Value(value).Build()
	}
	ttl := int64(expiration / time.Second)
	if ttl < 1 {
		ttl = 1
	}
	return c.rDB.B().Setex().Key(key).Seconds(ttl).Value(value).Build()
}
