package main

import (
	"net/http"
	"os"

	"github.com/nsarup/imgapi/internal/config"
	"github.com/nsarup/imgapi/internal/httpapi"
	"github.com/nsarup/imgapi/internal/logging"
	"github.com/nsarup/imgapi/internal/service"
	"github.com/nsarup/imgapi/internal/storage"
)

func main() {
	cfg := config.LoadFromEnv()
	log := logging.New(os.Stdout)

	store, err := storage.NewFileStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to init storage: %v", err)
	}
	svc := service.New(store)
	srv := httpapi.NewServer(cfg, log, svc)

	log.Printf("listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, srv.Handler()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
