package enrich

import (
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Redis implementation to store enrichment info
type redisImpl struct {
	pool *redis.Pool
}

// Initialize redis parameters
func NewRedisImpl(address, network string,
	maxIdleConn int,
	idleTimeout, connTimeout, readTimeout, writeTimeout time.Duration) DataInterface {

	p := &redis.Pool{
		MaxIdle:     maxIdleConn,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, address,
				redis.DialConnectTimeout(connTimeout),
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout))
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {

			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
	return &redisImpl{
		pool: p,
	}
}

// Set key and value into Redis
func (rd *redisImpl) Set(key string, value string) error {
	conn := rd.pool.Get()
	_, err := conn.Do("SET", key, value)
	if err != nil {
		//fmt.Println("Cannot set the value for", key, ":", value)
		return err
	}
	return nil

}

// Get value for specific key from  Redis
func (rd *redisImpl) Get(key string) (string, error) {
	conn := rd.pool.Get()
	s, err := redis.String(conn.Do("GET", key))
	if err != nil {
		//fmt.Println("value not found", key)
		return "", err
	}
	return s, nil
}

// Set keys and values in bulk into Redis
func (rd *redisImpl) SetBulk(pairs map[string]string) error {
	conn := rd.pool.Get()
	if pairs == nil {
		return errors.New("Nothing to load")
	}
	conn.Send("MULTI")
	for k, v := range pairs {
		conn.Send("SET", k, v)
	}
	_, err := conn.Do("EXEC")
	if err != nil {
		fmt.Println("Cannot set the values")
	}
	return err
}
