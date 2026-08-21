package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-benchmark/internal/config"
	"github.com/yourorg/go-benchmark/internal/handler"
	"github.com/yourorg/go-benchmark/internal/i18n"
	"github.com/yourorg/go-benchmark/internal/middleware"
	"github.com/yourorg/go-benchmark/internal/service"
	"github.com/yourorg/go-benchmark/internal/service/store"
	"github.com/yourorg/go-benchmark/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if err := i18n.EnsureLoaded(); err != nil {
		log.Fatalf("i18n: %v", err)
	}

	fs, err := store.New(cfg.ReportDir)
	if err != nil {
		log.Fatalf("report store: %v", err)
	}
	mgr := service.NewManager(cfg, fs)
	renderer, err := web.NewRenderer()
	if err != nil {
		log.Fatalf("web renderer: %v", err)
	}

	gin.SetMode(cfg.ServerMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.CORSAllowOrigins))

	web.RegisterStatic(r)
	h := handler.New(cfg, mgr, renderer)
	h.RegisterRoutes(r)

	log.Printf("go-benchmark listening on %s (lang=%s)", cfg.ServerAddr, cfg.DefaultLang)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
