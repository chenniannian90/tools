package confhttp

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/httpx"
	reflectx "github.com/go-courier/x/reflect"
)

type LivenessChecker interface {
	LivenessCheck() map[string]string
}

var checkers = livenessCheckers{}

func RegisterCheckerFromStruct(v interface{}) {
	rv := reflectx.Indirect(reflect.ValueOf(v))
	typ := rv.Type()

	if typ.Kind() != reflect.Struct {
		panic(errors.New("not struct"))
	}

	for i := 0; i < rv.NumField(); i++ {
		value := rv.Field(i)
		name := typ.Field(i).Name

		if livenessChecker, ok := value.Interface().(LivenessChecker); ok {
			RegisterChecker(name, livenessChecker)
		}
	}
}

func RegisterChecker(k string, checker LivenessChecker) {
	checkers[k] = checker
}

type livenessCheckers map[string]LivenessChecker

func (checkers livenessCheckers) Statuses() map[string]string {
	m := map[string]string{}

	for k := range checkers {
		if checkers[k] != nil {
			subStatuses := checkers[k].LivenessCheck()
			for subKey, v := range subStatuses {
				if subKey != "" {
					m[k+"/"+subKey] = v
				} else {
					m[k] = v
				}
			}
		}
	}

	return m
}

var LivenessRouter = courier.NewRouter(&Liveness{})

type Liveness struct {
	httpx.MethodGet
}

func (Liveness) Path() string {
	return "/liveness"
}

func (Liveness) Output(ctx context.Context) (interface{}, error) {
	return checkers.Statuses(), nil
}

type LivenessStatuses map[string]string
