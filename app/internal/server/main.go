package server

import (
	"localstack/internal/config"
	"net/http"
)

func StartHttpServer(cfg config.HTTPServerConfig, router http.Handler, shutdown chan bool) *http.Server {
	srv := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			close(shutdown)
		}
	}()

	return srv
}
