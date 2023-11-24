package middlewares

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

func RecoverHandler() func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return &recoverHandler{
			nextHandler: handler,
		}
	}
}

type recoverHandler struct {
	nextHandler http.Handler
}

func (h *recoverHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("recover err:%v\n", r)
		}
	}()
	h.nextHandler.ServeHTTP(rw, req)
}
