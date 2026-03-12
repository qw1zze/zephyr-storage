package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"zephyr-storage/internal/config"
	"zephyr-storage/internal/handler"
	"zephyr-storage/internal/ipfs"
	"zephyr-storage/internal/service"
)

func main() {
	cfg := config.Load()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ipfsClient := ipfs.NewClient(cfg.IPFSApiURL, cfg.IPFSGatewayURL, log)
	svc := service.NewService(ipfsClient, log)
	h := handler.NewHandler(svc, log)

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		BodyLimit:    int(cfg.MaxBlobSizeMB) * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(requestLogger(log))

	app.Get("/healthz", h.Health)
	app.Get("/readyz", h.Ready)

	v1 := app.Group("/v1")
	v1.Post("/blobs", h.UploadBlob)
	v1.Get("/blobs/:cid", h.GetBlob)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.ServerPort)
		log.Info("starting server", "addr", addr)
		if err := app.Listen(addr); err != nil {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = ctx
	if err := app.Shutdown(); err != nil {
		log.Error("shutdown error", "error", err)
	}

	log.Info("server stopped")
}

func requestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(start).String(),
		)
		return err
	}
}
