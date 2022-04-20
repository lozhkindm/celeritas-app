package cache

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	Conn   *redis.Pool
	Prefix string
}

func (rc *RedisCache) Has(key string) (bool, error) {
	key = rc.key(key)
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	found, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}
	return found, err
}

func (rc *RedisCache) Get(key string) (interface{}, error) {
	key = rc.key(key)
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	entry, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	decoded, err := decode(entry)
	if err != nil {
		return nil, err
	}

	return decoded[key], nil
}

func (rc *RedisCache) Set(key string, val interface{}, expires ...int) error {
	key = rc.key(key)
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	entry := Entry{}
	entry[key] = val
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		if _, err := conn.Do("SETEX", key, expires[0], string(encoded)); err != nil {
			return err
		}
	} else {
		if _, err := conn.Do("SET", key, string(encoded)); err != nil {
			return err
		}
	}

	return nil
}

func (rc *RedisCache) Forget(key string) error {
	key = rc.key(key)
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	if _, err := conn.Do("DEL", key); err != nil {
		return err
	}

	return nil
}

func (rc *RedisCache) Empty() error {
	return rc.EmptyByMatch("")
}

func (rc *RedisCache) EmptyByMatch(pattern string) error {
	pattern = rc.key(pattern)
	conn := rc.Conn.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	iter := 0
	keys := make([]string, 0)

	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", fmt.Sprintf("%s*", pattern)))
		if err != nil {
			return err
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	for _, key := range keys {
		if _, err := conn.Do("DEL", key); err != nil {
			return err
		}
	}

	return nil
}

func (rc *RedisCache) key(key string) string {
	return fmt.Sprintf("%s:%s", rc.Prefix, key)
}
