package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"life_easy_api/api/controllers/life"
	"life_easy_api/api/middleware"
	lifeweb "life_easy_api/api/routes/life-portal-web"
	"life_easy_api/api/services"
	"life_easy_api/config"
	dbpkg "life_easy_api/pkg/db-pkg"
	applogger "life_easy_api/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		client := http.Client{Timeout: 2 * time.Second}

		response, err := client.Get(
			"http://127.0.0.1:8080/health/ready",
		)
		if err != nil {
			os.Exit(1)
		}

		_ = response.Body.Close()

		if response.StatusCode != http.StatusOK {
			os.Exit(1)
		}

		return
	}

	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, err := applogger.New()
	if err != nil {
		return err
	}
	defer logger.Sync()

	db, err := dbpkg.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return errors.New("database startup failed")
	}
	defer db.Close()

	service := &services.Service{
		DB:        db,
		JWTSecret: cfg.JWTSecret,
	}

	controller := &life.Controller{Service: service}

	app := fiber.New(fiber.Config{
		AppName:               "Life Easy API",
		Immutable:             true,
		BodyLimit:             1024 * 1024,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		DisableStartupMessage: true,
		ErrorHandler:          middleware.ErrorHandler,
	})

	app.Use(requestid.New())
	app.Use(middleware.AccessLog(logger))
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(middleware.QueryTimeout())

	app.Get("/health/live", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/health/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			time.Second,
		)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return c.Status(503).JSON(
				fiber.Map{"status": "not_ready"},
			)
		}

		return c.JSON(fiber.Map{"status": "ready"})
	})

	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		c.Type("yaml")
		return c.SendFile("./docs/openapi/life_web.yaml")
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline'; "+
				"font-src 'self' data: https:; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self' https:;")

		c.Type("html")

		return c.SendString(`<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Life Easy API</title>
</head>
<body>
  <div id="app"></div>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  <script>
    Scalar.createApiReference('#app', {
      url: '/openapi.yaml'
    });
  </script>
</body>
</html>`)
	})

	lifeweb.Register(app, controller, cfg.JWTSecret)

	serverError := make(chan error, 1)
	go func() {
		serverError <- app.Listen(":8080")
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-serverError:
		return err

	case <-signals:
		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		return app.ShutdownWithContext(ctx)
	}
}
