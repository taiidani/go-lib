package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Redis struct {
	client   *redis.Client
	dbPrefix string
}

var _ Cache = &Redis{}

const (
	defaultRedisPort = "6379"
	defaultRedisDB   = 0
)

// NewRedis instantiates a new Cache client backed by Redis.
// This will connect to Redis based on the presence of environment variables.
//
// If the `REDIS_USER` or `REDIS_PASSWORD` (or `REDIS_PASS`) environment variables
// are set, the connection will be upgraded to secure.
//
// By default DB 0 is used, but can be configured with the `REDIS_DB` environment
// variable.
func NewRedis(dbPrefix string) (*Redis, error) {
	host, port, user, password, db := parseEnvVars()

	if len(user) > 0 || len(password) > 0 {
		return NewRedisSecure(host, port, user, password, db, dbPrefix)
	}

	return NewRedisInsecure(host, port, db, dbPrefix)
}

func parseEnvVars() (host, port, user, password string, db int) {
	host = os.Getenv("REDIS_HOST")
	port = os.Getenv("REDIS_PORT")
	if len(port) == 0 {
		port = defaultRedisPort
	}
	if addr := os.Getenv("REDIS_ADDR"); len(addr) > 0 {
		components := strings.Split(addr, ":")
		if len(components) == 1 {
			host = components[0]
			port = defaultRedisPort
		} else if len(components) == 2 {
			host = components[0]
			port = components[1]
		}
	}

	user = os.Getenv("REDIS_USER")
	password = os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = os.Getenv("REDIS_PASS")
	}

	db = defaultRedisDB
	if dbString, ok := os.LookupEnv("REDIS_DB"); ok {
		db, _ = strconv.Atoi(dbString)
	}

	return host, port, user, password, db
}

// NewRedisInsecure instantiates a new insecure client
func NewRedisInsecure(host, port string, db int, dbPrefix string) (*Redis, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
		DB:   db,
	})
	status := client.Ping(ctx)
	if status.Err() != nil {
		return nil, fmt.Errorf("failed to create redis client %q: %w", client.Options().Addr, status.Err())
	}

	return &Redis{client: client, dbPrefix: dbPrefix}, nil
}

// NewRedisSecure instantiates a new secure TLS client
func NewRedisSecure(host, port, user, password string, db int, dbPrefix string) (*Redis, error) {
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
