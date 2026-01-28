// Package server provides the Fiber-based HTTP engine for Insta-Mock.
// It dynamically generates REST API routes from a JSON data structure.
package server

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

// Engine holds the Fiber app and in-memory data store.
type Engine struct {
	app   *fiber.App
	store map[string][]map[string]interface{}
	mu    sync.RWMutex
}

// NewEngine creates a new Engine instance with dynamic routes based on the provided data.
// It accepts a generic map[string]interface{} representing the loaded JSON.
// Each top-level key becomes a REST resource with full CRUD support.
func NewEngine(data map[string]interface{}) *Engine {
	e := &Engine{
		app: fiber.New(fiber.Config{
			AppName:               "Insta-Mock",
			DisableStartupMessage: true,
		}),
		store: make(map[string][]map[string]interface{}),
	}

	// Enable CORS for all origins
	e.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Normalize input data: ensure each key maps to a slice
	e.normalizeData(data)

	// Register dynamic routes for each resource
	e.registerRoutes()

	return e
}

// normalizeData converts the input JSON into slices for consistent handling.
// If a key contains an object, it wraps it in a slice.
// If it contains a slice, it converts each element to map[string]interface{}.
func (e *Engine) normalizeData(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case []interface{}:
			// Convert []interface{} to []map[string]interface{}
			items := make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					// Ensure each item has an ID
					if _, hasID := m["id"]; !hasID {
						m["id"] = uuid.New().String()
					}
					items = append(items, m)
				}
			}
			e.store[key] = items
		case map[string]interface{}:
			// Wrap single object in a slice
			if _, hasID := v["id"]; !hasID {
				v["id"] = uuid.New().String()
			}
			e.store[key] = []map[string]interface{}{v}
		default:
			// Skip non-object/array values
			continue
		}
	}
}

// registerRoutes dynamically creates CRUD endpoints for each resource in the store.
func (e *Engine) registerRoutes() {
	for resource := range e.store {
		res := resource // capture for closures

		// GET /:resource - List all items
		e.app.Get("/"+res, e.handleGetAll(res))

		// GET /:resource/:id - Get single item by ID
		e.app.Get("/"+res+"/:id", e.handleGetByID(res))

		// POST /:resource - Create new item
		e.app.Post("/"+res, e.handleCreate(res))

		// PUT /:resource/:id - Update existing item
		e.app.Put("/"+res+"/:id", e.handleUpdate(res))

		// DELETE /:resource/:id - Delete item
		e.app.Delete("/"+res+"/:id", e.handleDelete(res))
	}

	// Health check endpoint
	e.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"resources": e.listResources(),
		})
	})
}

// listResources returns a list of available resource names.
func (e *Engine) listResources() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	resources := make([]string, 0, len(e.store))
	for key := range e.store {
		resources = append(resources, key)
	}
	return resources
}

// handleGetAll returns a handler that lists all items for a resource.
func (e *Engine) handleGetAll(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		e.mu.RLock()
		defer e.mu.RUnlock()

		items := e.store[resource]
		return c.JSON(items)
	}
}

// handleGetByID returns a handler that retrieves a single item by ID.
func (e *Engine) handleGetByID(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		e.mu.RLock()
		defer e.mu.RUnlock()

		for _, item := range e.store[resource] {
			if itemID, ok := item["id"]; ok && fmt.Sprintf("%v", itemID) == id {
				return c.JSON(item)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":    "not_found",
			"message":  fmt.Sprintf("%s with id '%s' not found", resource, id),
			"resource": resource,
		})
	}
}

// handleCreate returns a handler that creates a new item.
func (e *Engine) handleCreate(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_body",
				"message": "Request body must be valid JSON",
			})
		}

		// Auto-generate ID if not provided
		if _, hasID := body["id"]; !hasID {
			body["id"] = uuid.New().String()
		}

		e.mu.Lock()
		e.store[resource] = append(e.store[resource], body)
		e.mu.Unlock()

		return c.Status(fiber.StatusCreated).JSON(body)
	}
}

// handleUpdate returns a handler that updates an existing item.
func (e *Engine) handleUpdate(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_body",
				"message": "Request body must be valid JSON",
			})
		}

		e.mu.Lock()
		defer e.mu.Unlock()

		for i, item := range e.store[resource] {
			if itemID, ok := item["id"]; ok && fmt.Sprintf("%v", itemID) == id {
				// Preserve ID, update other fields
				body["id"] = itemID
				e.store[resource][i] = body
				return c.JSON(body)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":    "not_found",
			"message":  fmt.Sprintf("%s with id '%s' not found", resource, id),
			"resource": resource,
		})
	}
}

// handleDelete returns a handler that removes an item by ID.
func (e *Engine) handleDelete(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		e.mu.Lock()
		defer e.mu.Unlock()

		items := e.store[resource]
		for i, item := range items {
			if itemID, ok := item["id"]; ok && fmt.Sprintf("%v", itemID) == id {
				// Remove item from slice
				e.store[resource] = append(items[:i], items[i+1:]...)
				return c.Status(fiber.StatusNoContent).Send(nil)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":    "not_found",
			"message":  fmt.Sprintf("%s with id '%s' not found", resource, id),
			"resource": resource,
		})
	}
}

// Start runs the Fiber server on the specified address.
func (e *Engine) Start(addr string) error {
	return e.app.Listen(addr)
}

// Shutdown gracefully stops the Fiber server.
func (e *Engine) Shutdown() error {
	return e.app.Shutdown()
}

// App returns the underlying Fiber app for advanced configuration.
func (e *Engine) App() *fiber.App {
	return e.app
}
