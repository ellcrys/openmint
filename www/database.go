package www

import (
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"time"
)

// Establish a mongo db connection
func GetMongoSession(host, database, username, password string) (*mgo.Session, error) {
	return mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{host},
		Timeout:  5 * time.Second,
		Database: database,
		Username: username,
		Password: password,
	})
}

// Establish a redis connection
func GetRedisPool(addr, password string, db int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr, redis.DialDatabase(db), redis.DialPassword(password))
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
