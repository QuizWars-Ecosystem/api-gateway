package mux

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
)

func RegisterRuntimeMux(mux *runtime.ServeMux) error {
	var err error

	if err = mux.HandlePath("GET", "/healthz", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}); err != nil {
		return err
	}

	return nil
}
