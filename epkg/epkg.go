package epkg

import (
	"fmt"
	"github.com/go-courier/statuserror"
	"github.com/pkg/errors"
)

func Wrap(err error, message string) error {
	if statusErr, ok := statuserror.IsStatusErr(err); ok {
		statusErr.Desc = message + ": " + statusErr.Desc

		return statusErr
	}

	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	if statusErr, ok := statuserror.IsStatusErr(err); ok {
		prefix := fmt.Sprintf(format, args...) + ": "
		statusErr.Desc = prefix + statusErr.Desc

		return statusErr
	}

	return errors.Wrapf(err, format, args...)
}

func WithMessage(err error, message string) error {
	if statusErr, ok := statuserror.IsStatusErr(err); ok {
		statusErr.Desc = message + ": " + statusErr.Desc

		return statusErr
	}

	return errors.WithMessage(err, message)
}

func WithMessagef(err error, format string, args ...interface{}) error {
	// v, ok := err.(e.StatusError)
	// if ok {
	//	code := v.ServiceCode()
	//
	//	if IntListContain(RemainErrorCodeList, code) || code/1e6 == http.StatusConflict {
	//		return err
	//	}
	// }
	if statusErr, ok := statuserror.IsStatusErr(err); ok {
		prefix := fmt.Sprintf(format, args...) + ": "
		statusErr.Desc = prefix + statusErr.Desc

		return statusErr
	}

	return errors.WithMessagef(err, format, args...)
}

func IntListContain(list []int, val int) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}

func As(err error, target interface{}) bool { return errors.As(err, target) }
