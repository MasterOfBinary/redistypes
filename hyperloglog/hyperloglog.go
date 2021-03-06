// Package hyperloglog contains a Go implementation of the HyperLogLog data structure in Redis.
//
// For more information about how the data structure works, see the Redis documentation or http://antirez.com/news/75.
package hyperloglog

import (
	"github.com/MasterOfBinary/redistypes"
	"github.com/MasterOfBinary/redistypes/internal"
	"github.com/garyburd/redigo/redis"
)

// HyperLogLog is a probabilistic data structure that counts the number of unique items
// added to it.
type HyperLogLog interface {
	// Base returns the base Type.
	Base() redistypes.Type

	// Add implements the Redis command PFADD. It adds items to the HyperLogLog count. It returns an error or true
	// if at least one internal register was altered, or false otherwise.
	//
	// See https://redis.io/commands/pfadd.
	Add(args ...interface{}) (bool, error)

	// Count implements the Redis command PFCOUNT. It returns the count of unique items added to the HyperLogLog,
	// or an error if something went wrong.
	//
	// See https://redis.io/commands/pfcount.
	Count() (uint64, error)

	// Merge implements the Redis command PFMERGE. It merges the HyperLogLog with other to produce a new
	// HyperLogLog with given name. It returns an error or the newly created HyperLogLog.
	//
	// See https://redis.io/commands/pfmerge.
	Merge(name string, other HyperLogLog) (HyperLogLog, error)
}

type redisHyperLogLog struct {
	conn redis.Conn
	base redistypes.Type
}

// NewRedisHyperLogLog creates a Redis implementation of HyperLogLog given redigo connection conn and name. The
// Redis key used to identify the HyperLogLog will be name.
func NewRedisHyperLogLog(conn redis.Conn, name string) HyperLogLog {
	return &redisHyperLogLog{
		conn: conn,
		base: redistypes.NewRedisType(conn, name),
	}
}

func (r redisHyperLogLog) Base() redistypes.Type {
	return r.base
}

func (r redisHyperLogLog) Add(args ...interface{}) (bool, error) {
	args = internal.PrependInterface(r.base.Name(), args...)
	return redis.Bool(r.conn.Do("PFADD", args...))
}

func (r *redisHyperLogLog) Count() (uint64, error) {
	return redis.Uint64(r.conn.Do("PFCOUNT", r.base.Name()))
}

func (r *redisHyperLogLog) Merge(name string, other HyperLogLog) (HyperLogLog, error) {
	_, err := redis.String(r.conn.Do("PFMERGE", name, r.base.Name(), other.Base().Name()))

	if err != nil {
		return nil, err
	}

	return NewRedisHyperLogLog(r.conn, name), nil
}
