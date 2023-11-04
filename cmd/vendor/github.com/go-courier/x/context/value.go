package context

import (
	"context"
	"fmt"
	"reflect"
)

// WithValue seem as context.WithValue but without key type comparable check, and be 20x faster
func WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	if key == nil {
		panic("nil key")
	}
	return &valueCtx{parent, key, val}
}

type valueCtx struct {
	context.Context
	key, val interface{}
}

func (c *valueCtx) String() string {
	return contextName(c.Context) + ".WithValue(type " + reflect.TypeOf(c.key).String() + ", val " + stringify(c.val) + ")"
}

func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}

func stringify(v interface{}) string {
	switch s := v.(type) {
	case fmt.Stringer:
		return s.String()
	case string:
		return s
	}
	return "<not Stringer>"
}

func contextName(c context.Context) string {
	if s, ok := c.(fmt.Stringer); ok {
		return s.String()
	}
	return reflect.TypeOf(c).String()
}
