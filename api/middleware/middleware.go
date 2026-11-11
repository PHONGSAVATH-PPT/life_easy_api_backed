package middleware

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"life_easy_api/pkg/db-pkg/models"
	"life_easy_api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

func Auth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := strings.Fields(c.Get("Authorization"))

		if len(header) != 2 ||
			!strings.EqualFold(header[0], "Bearer") {
			return models.ErrUnauthorized
		}

		userID, err := utils.ParseToken(header[1], secret)
		if err != nil {
			return models.ErrUnauthorized
		}

		if _, err := uuid.Parse(userID); err != nil {
			return models.ErrUnauthorized
		}

		c.Locals("userID", userID)

		return c.Next()
	}
}

func UserID(c *fiber.Ctx) string {
	return c.Locals("userID").(string)
}

func QueryTimeout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		c.SetUserContext(ctx)

		return c.Next()
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "Internal server error"

	var fiberError *fiber.Error
	var pgError *pgconn.PgError

	switch {
	case errors.Is(err, models.ErrInvalid):
		status, message = 400, "Invalid request"
	case errors.Is(err, models.ErrUnauthorized):
		status, message = 401, "Invalid or expired credentials"
	case errors.Is(err, models.ErrForbidden):
		status, message = 403, "Permission denied"
	case errors.Is(err, sql.ErrNoRows):
		status, message = 404, "Resource not found or not editable"
	case errors.Is(err, context.DeadlineExceeded):
		status, message = 504, "Request timed out"
	case errors.As(err, &pgError):
		switch pgError.Code {
		case "23505":
			status, message = 409, "Resource already exists"
		case "23503", "23514", "22P02", "22001":
			status, message = 400, "Invalid related resource or value"
		}
	case errors.As(err, &fiberError):
		status = fiberError.Code
		if status < 500 {
			message = fiberError.Message
		}
	}

	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"message":    message,
			"request_id": c.GetRespHeader("X-Request-ID"),
		},
	})
}

func AccessLog(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		if err != nil {
			if writeErr := ErrorHandler(c, err); writeErr != nil {
				return writeErr
			}
		}

		// ไม่บันทึก password, token, request body หรือ query string
		log.Info("http_request",
			zap.String("request_id", c.GetRespHeader("X-Request-ID")),
			zap.String("method", c.Method()),
			zap.String("route", c.Route().Path),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", time.Since(start)),
		)

		return nil
	}
}
