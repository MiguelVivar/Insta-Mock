package server

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

// chaosMiddleware introduces random failures and latency for testing.
func chaosMiddleware(failPercent int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Random latency (50-500ms)
		delay := time.Duration(50+rand.Intn(450)) * time.Millisecond
		time.Sleep(delay)

		// Random failure
		if rand.Intn(100) < failPercent {
			errors := []int{
				fiber.StatusInternalServerError,
				fiber.StatusBadGateway,
				fiber.StatusServiceUnavailable,
				fiber.StatusGatewayTimeout,
			}
			status := errors[rand.Intn(len(errors))]
			return c.Status(status).JSON(fiber.Map{
				"error":   "chaos_error",
				"message": "Simulated failure from chaos mode",
				"status":  status,
			})
		}

		return c.Next()
	}
}
