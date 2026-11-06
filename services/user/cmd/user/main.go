// Composition root: único lugar que instancia adapters e os conecta aos casos de uso.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	"q-auction/user/internal/adapters/in/httpapi"
	"q-auction/user/internal/config"
)

func main() {
	cfg := config.Load()

	app := httpapi.NewApp(cfg.Service + "-service")
	httpapi.NewHealthHandler(cfg.Service, cfg.Instance).Register(app)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":"+cfg.Port, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()
	log.Printf("%s-service ouvindo na porta %s", cfg.Service, cfg.Port)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
