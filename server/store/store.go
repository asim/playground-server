package store

import (
	"errors"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	zprefix = "zrange:"
	store   *redis.Pool
)

var (
	ErrNotFound = errors.New("Not Found")
)

func initStore() {
	if store != nil {
		return
	}

	host := os.Getenv("PLAYGROUND_REDIS_SERVICE_HOST")
	port := os.Getenv("PLAYGROUND_REDIS_SERVICE_PORT")

	if len(host) == 0 {
		host = "127.0.0.1"
	}

	if len(port) == 0 {
		port = "6379"
	}

	store = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", host+":"+port)
	}, 5)
}

func Exists(namespace, key string) (bool, error) {
	initStore()
	conn := store.Get()
	defer conn.Close()
	return redis.Bool(conn.Do("HEXISTS", namespace, key))
}

func Get(namespace, key string) ([]byte, error) {
	initStore()
	conn := store.Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("HGET", namespace, key))
	if err == redis.ErrNil {
		return nil, ErrNotFound
	}
	return b, err
}

func Put(namespace, key string, value []byte) error {
	initStore()
	conn := store.Get()
	defer conn.Close()

	_, err := conn.Do("ZADD", zprefix+namespace, time.Now().UnixNano(), key)
	if err != nil {
		return err
	}
	_, err = conn.Do("HSET", namespace, key, value)
	return err
}

func Del(namespace, key string) error {
	initStore()
	conn := store.Get()
	defer conn.Close()

	_, err := conn.Do("ZREM", zprefix+namespace, key)
	if err != nil {
		return err
	}
	_, err = conn.Do("HDEL", namespace, key)
	return err
}

func Range(namespace string, offset, limit int) ([][]byte, error) {
	// lookup zrange
	// get keys with hmget
	initStore()
	conn := store.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("ZREVRANGE", zprefix+namespace, offset, limit))
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return [][]byte{}, nil
	}

	args := []interface{}{namespace}
	for _, key := range keys {
		args = append(args, key)
	}

	result, err := redis.Strings(conn.Do("HMGET", args...))
	if err != nil {
		return nil, err
	}

	var results [][]byte
	for _, res := range result {
		results = append(results, []byte(res))
	}

	return results, nil
}
