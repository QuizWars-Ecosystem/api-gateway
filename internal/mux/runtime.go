package mux

import (
	"net/http"
)

func RegisterRuntimeMux(mux *http.ServeMux) error {

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return nil
}
