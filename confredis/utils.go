package confredis

import "github.com/gomodule/redigo/redis"

type Conn = redis.Conn

func Command(name string, args ...interface{}) *CMD {
	return &CMD{
		name: name,
		args: args,
	}
}

type CMD struct {
	name string
	args []interface{}
}

type RedisOperator interface {
	// Prefix key prefix
	Prefix(key string) string
	// Get get redis connect
	Get() Conn

	// Exec exec
	Exec(cmd *CMD, others ...*CMD) (interface{}, error)
}
