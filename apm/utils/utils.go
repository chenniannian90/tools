package utils

import (
	"reflect"
	"strings"
)

func NormOperator(operator string) string {
	idx := strings.LastIndex(operator, "#")
	if 0 <= idx && idx < len(operator) {
		return operator[idx+1:]
	}
	idx = strings.LastIndex(operator, "/")
	if 0 <= idx && idx < len(operator) {
		return operator[idx+1:]
	}

	idx = strings.LastIndex(operator, ".")
	if 0 <= idx && idx < len(operator) {
		return operator[idx+1:]
	}
	return operator
}

func CalcOperator(req interface{}) string {
	rt := reflect.TypeOf(req)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	return rt.Name()
}
