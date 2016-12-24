package redistypes

import (
	"github.com/garyburd/redigo/redis"
	"github.com/golang/groupcache/singleflight"
)

// HyperLogLog is a probabilistic data structure that counts the number of unique items
// added to it. For more information about how the data structure works, see the Redis
// documentation or http://antirez.com/news/75.
type HyperLogLog interface {
	// Name returns the name of the HyperLogLog.
	Name() string

	// Add adds items to the HyperLogLog count. It returns an error or true if at
	// least one internal register was altered, or false otherwise.
	//
	// See https://redis.io/commands/pfadd.
	Add(args ...interface{}) (bool, error)

	// Count returns the count of unique items added to the HyperLogLog, or an
	// error if something went wrong.
	//
	// See https://redis.io/commands/pfcount.
	Count() (uint64, error)

	// Merge merges the HyperLogLog with other to produce a new HyperLogLog with
	// a given name. It returns an error or the newly created HyperLogLog.
	//
	// See https://redis.io/commands/pfmerge.
	Merge(name string, other HyperLogLog) (HyperLogLog, error)
}

type redisHyperLogLog struct {
	conn redis.Conn
	name string
	sync singleflight.Group
}

// NewRedisHyperLogLog creates a Redis implementation of HyperLogLog.
func NewRedisHyperLogLog(conn redis.Conn, name string) HyperLogLog {
	return &redisHyperLogLog{
		conn: conn,
		name: name,
	}
}

func (r redisHyperLogLog) Name() string {
	return r.name
}

func (r *redisHyperLogLog) Add(args ...interface{}) (bool, error) {
	args = prependInterface(r.name, args...)
	return redis.Bool(r.conn.Do("PFADD", args...))
}

func (r *redisHyperLogLog) Count() (uint64, error) {
	return redis.Uint64(r.sync.Do("PFCOUNT", func() (interface{}, error) {
		return r.conn.Do("PFCOUNT", r.name)
	}))
}

func (r *redisHyperLogLog) Merge(name string, other HyperLogLog) (HyperLogLog, error) {
	args := make([]interface{}, 3)
	args[0] = name
	args[1] = r.name
	args[2] = other.Name()
	_, err := redis.String(r.sync.Do("PFMERGE:"+name+":"+other.Name(), func() (interface{}, error) {
		return r.conn.Do("PFMERGE", args...)
	}))

	if err != nil {
		return nil, err
	}

	return NewRedisHyperLogLog(r.conn, name), nil
}
