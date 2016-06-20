package www

import (
	"time"
	"gopkg.in/mgo.v2"
	"github.com/garyburd/redigo/redis"
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
func GetRedisConnection(addr, password string, db int) (redis.Conn, error) {
	return redis.Dial("tcp", addr, redis.DialDatabase(db), redis.DialPassword(password))
}