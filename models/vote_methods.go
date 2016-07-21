package models

import (
	"errors"
	"fmt"
	"github.com/ellcrys/openmint/config"
	"github.com/garyburd/redigo/redis"
)

// The name of the redis list acting as a queue
var VOTE_QUEUE_NAME = "openmint_vote_queue"

// The vote session prefix
var VOTE_SESSION_PREFIX = "openmint_vote_session"

// Adds a currency id to the redis list
// being used as a queue.
func AddToVoteQueue(redisPool *redis.Pool, currencyId string) error {
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := redis.Int64(conn.Do("RPUSH", VOTE_QUEUE_NAME, currencyId)); err != nil {
		return err
	}
	return nil
}

// Gets a currency from the tail of vote queue and also
// move the returned currency to the head of the vote queue
// using the RPOPLPUSH command
func GetFromVoteQueue(redisPool *redis.Pool) (string, error) {
	conn := redisPool.Get()
	defer conn.Close()
	currencyId, err := redis.String(conn.Do("RPOPLPUSH", VOTE_QUEUE_NAME, VOTE_QUEUE_NAME))
	if err != nil {
		return "", err
	}
	return currencyId, nil
}

// Remove a currency id from the vote queue list
func RemoveFromVoteQueue(redisPool *redis.Pool, currencyId string) error {
	conn := redisPool.Get()
	defer conn.Close()
	if _, err := redis.Int64(conn.Do("LREM", VOTE_QUEUE_NAME, 0, currencyId)); err != nil {
		return err
	}
	return nil
}

// Count the number of active session
func CountActiveSessions(redisPool *redis.Pool, currencyId string) (int, error) {

	// var currencySessionKey = VOTE_SESSION_PREFIX + "_" + currencyId
	var currencySessionKey = "myset"
	conn := redisPool.Get()
	defer conn.Close()

	// get active session list
	voteSessionList, err := redis.Values(conn.Do("SMEMBERS", currencySessionKey))
	if err != nil {
		return 0, err
	}

	if len(voteSessionList) == 0 {
		return 0, nil
	}

	// get the individual session keys
	sessionKeys := []interface{}{}
	for _, sk := range voteSessionList {
		sessionKeys = append(sessionKeys, "vote_session_"+string(sk.([]byte)))
	}

	// count how many of the session keys are still alive (not expired)
	sessionState, err := redis.Values(conn.Do("MGET", sessionKeys...))
	if err != nil {
		return 0, err
	}

	aliveCount := 0
	for _, state := range sessionState {
		if state != nil {
			aliveCount++
		}
	}

	return aliveCount, nil
}

// Add a new session to a currency session set
func AddNewSession(redisPool *redis.Pool, currencyId, voteSessionId string) error {

	// var currencySessionKey = VOTE_SESSION_PREFIX + "_" + currencyId
	var currencySessionKey = "myset"
	var voteSessionKey = "vote_session_" + voteSessionId
	conn := redisPool.Get()
	defer conn.Close()

	// get session list
	_, err := redis.Int64(conn.Do("SADD", currencySessionKey, voteSessionId))
	if err != nil {
		return err
	}

	// set EXPIRE time to currency session (we don't want it living for ever)
	_, err = redis.Int64(conn.Do("EXPIRE", currencySessionKey, 60*30))
	if err != nil {
		return err
	}

	// set session id and it's expiry time. This indicates the
	// session is active for the giving time before it's expiry time
	_, err = redis.String(conn.Do("SETEX", voteSessionKey, config.C.GetInt("vote_session_duration"), "-"))
	if err != nil {
		return err
	}

	return nil
}

// Checks if a vote session is valid and active
func IsActiveSession(redisPool *redis.Pool, currencyId, voteSessionId string) (bool, error) {

	var validVoteSessionId = false

	// var currencySessionKey = VOTE_SESSION_PREFIX + "_" + currencyId
	var currencySessionKey = "myset"
	conn := redisPool.Get()
	defer conn.Close()

	// get active session list
	voteSessionList, err := redis.Values(conn.Do("SMEMBERS", currencySessionKey))
	if err != nil {
		return false, err
	}

	// ensure vote session id exists in active session list
	for _, sk := range voteSessionList {
		if string(sk.([]byte)) == voteSessionId {
			validVoteSessionId = true
		}
	}

	if !validVoteSessionId {
		fmt.Println("Vote session id is not known")
		return false, errors.New("vote session id is unknown")
	}

	// check if the vote session is still active
	_, err = redis.String(conn.Do("GET", "vote_session_"+voteSessionId))
	if err != nil && err == redis.ErrNil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
