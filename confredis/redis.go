package confredis

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-courier/envconf"
	"github.com/gomodule/redigo/redis"
)

type Redis struct {
	Protocol       string
	Host           string `env:",upstream"`
	Port           int
	Password       envconf.Password `env:""`
	ConnectTimeout envconf.Duration
	ReadTimeout    envconf.Duration
	WriteTimeout   envconf.Duration
	IdleTimeout    envconf.Duration
	MaxActive      int
	MaxIdle        int
	Wait           bool
	DB             int
	pool           *redis.Pool
}

var prefix = ""

func (r *Redis) Get() Conn {
	if r.pool != nil {
		return r.pool.Get()
	}
	return nil
}

func (r *Redis) Exec(cmd *CMD, others ...*CMD) (interface{}, error) {
	c := r.Get()
	defer c.Close()

	if (len(others)) == 0 {
		return c.Do(cmd.name, cmd.args...)
	}

	err := c.Send("MULTI")
	if err != nil {
		return nil, err
	}

	err = c.Send(cmd.name, cmd.args...)
	if err != nil {
		return nil, err
	}

	for i := range others {
		o := others[i]
		if o == nil {
			continue
		}
		err := c.Send(o.name, o.args...)
		if err != nil {
			return nil, err
		}
	}

	return c.Do("EXEC")
}

func (r *Redis) Prefix(key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

func (r *Redis) LivenessCheck() map[string]string {
	m := map[string]string{}

	conn := r.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		m[r.Host] = err.Error()
	} else {
		m[r.Host] = "ok"
	}

	return m
}

func (r *Redis) SetDefaults() {
	if r.Protocol == "" {
		r.Protocol = "tcp"
	}
	if r.Port == 0 {
		r.Port = 6379
	}
	if r.ConnectTimeout == 0 {
		r.ConnectTimeout = envconf.Duration(10 * time.Second)
	}
	if r.ReadTimeout == 0 {
		r.ReadTimeout = envconf.Duration(10 * time.Second)
	}
	if r.WriteTimeout == 0 {
		r.WriteTimeout = envconf.Duration(10 * time.Second)
	}
	if r.IdleTimeout == 0 {
		r.IdleTimeout = envconf.Duration(240 * time.Second)
	}
	if r.MaxActive == 0 {
		r.MaxActive = 5
	}
	if r.MaxIdle == 0 {
		r.MaxIdle = 3
	}
	if !r.Wait {
		r.Wait = true
	}
	if r.DB == 0 {
		r.DB = 10
	}
}

func (r *Redis) Init() {
	if r.pool == nil {
		r.initial()
	}

	fmt.Println("RedisEndpoint init")
	var env = strings.ToLower(os.Getenv("GOENV"))
	var projectName = strings.ToLower(os.Getenv("PROJECT_NAME"))
	prefix = fmt.Sprintf("%s:%s", env, projectName)
}

func (r *Redis) initial() {
	dialFunc := func() (c redis.Conn, err error) {
		c, err = redis.Dial(
			r.Protocol,
			fmt.Sprintf("%s:%d", r.Host, r.Port),

			redis.DialWriteTimeout(time.Duration(r.WriteTimeout)),
			redis.DialConnectTimeout(time.Duration(r.ConnectTimeout)),
			redis.DialReadTimeout(time.Duration(r.ReadTimeout)),
			redis.DialPassword(r.Password.String()),
			redis.DialDatabase(r.DB),
		)
		return
	}

	r.pool = &redis.Pool{
		Dial:        dialFunc,
		MaxIdle:     r.MaxIdle,
		MaxActive:   r.MaxActive,
		IdleTimeout: time.Duration(r.IdleTimeout),
		Wait:        true,
	}
}
