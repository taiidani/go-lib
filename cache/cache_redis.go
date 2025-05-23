package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Redis struct {
	client   *redis.Client
	dbPrefix string
}

var _ Cache = &Redis{}

// NewRedis instantiates a new client
func NewRedis(addr string, dbPrefix string) (*Redis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := redis.NewClient(&redis.Options{Addr: addr})
	status := client.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("failed to create redis client %q: %w", client.Options().Addr, status.Err())
	}

	return &Redis{client: client, dbPrefix: dbPrefix}, nil
}

// NewRedisSecureCache instantiates a new secure TLS client
func NewRedisSecureCache(host, port, user, password string, db int, dbPrefix string) (*Redis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:      fmt.Sprintf("%s:%s", host, port),
		Username:  user,
		Password:  password,
		TLSConfig: &tls.Config{},
		DB:        db,
	})
	status := client.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("failed to create secure Redis client %q: %w", client.Options().Addr, status.Err())
	}

	return &Redis{client: client, dbPrefix: dbPrefix}, nil
}

func (c *Redis) Get(ctx context.Context, key string, val any) error {
	resp := c.client.Get(ctx, c.dbPrefix+key)
	if resp.Err() != nil {
		return resp.Err()
	}

	return json.Unmarshal([]byte(resp.Val()), val)
}

func (c *Redis) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	req, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, c.dbPrefix+key, req, ttl).Err()
}

func (c *Redis) Has(ctx context.Context, key string) (bool, error) {
	resp := c.client.Exists(ctx, c.dbPrefix+key)
	if resp.Err() != nil {
		return false, resp.Err()
	}

	return resp.Val() > 0, nil
}

func (c *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	resp := c.client.Keys(ctx, c.dbPrefix+pattern)
	if resp.Err() != nil {
		return []string{}, resp.Err()
	}

	return resp.Val(), nil
}
