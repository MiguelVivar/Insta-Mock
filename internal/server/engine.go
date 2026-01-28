// Package server provides the Fiber-based HTTP engine for Insta-Mock.
// It dynamically generates REST API routes from a JSON data structure.
package server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
)

// RequestLog represents a logged API request for TUI display.
type RequestLog struct {
	Method     string
	Path       string
	StatusCode int
	Latency    string
}

// Engine holds the Fiber app and in-memory data store.
type Engine struct {
	app       *fiber.App
	store     map[string][]map[string]interface{}
	mu        sync.RWMutex
	OnRequest func(log RequestLog) // Callback for TUI logging
}

// EngineConfig holds configuration options for the engine.
type EngineConfig struct {
	EnableLogger bool
	ChaosMode    bool
	ChaosPercent int // Percentage of requests to fail (0-100)
}

// NewEngine creates a new Engine instance with dynamic routes based on the provided data.
func NewEngine(data map[string]interface{}) *Engine {
	return NewEngineWithConfig(data, EngineConfig{})
}

// NewEngineWithConfig creates a new Engine with custom configuration.
func NewEngineWithConfig(data map[string]interface{}, config EngineConfig) *Engine {
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
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Optional request logger
	if config.EnableLogger {
		e.app.Use(logger.New(logger.Config{
			Format:     "${time} │ ${status} │ ${latency} │ ${method} ${path}\n",
			TimeFormat: "15:04:05",
		}))
	}

	// Chaos middleware
	if config.ChaosMode {
		e.app.Use(chaosMiddleware(config.ChaosPercent))
	}

	// Normalize input data
	e.normalizeData(data)

	// Register dynamic routes
	e.registerRoutes()

	return e
}

// normalizeData converts the input JSON into slices for consistent handling.
func (e *Engine) normalizeData(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case []interface{}:
			items := make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					if _, hasID := m["id"]; !hasID {
						m["id"] = uuid.New().String()
					}
					items = append(items, m)
				}
			}
			e.store[key] = items
		case map[string]interface{}:
			if _, hasID := v["id"]; !hasID {
				v["id"] = uuid.New().String()
			}
			e.store[key] = []map[string]interface{}{v}
		default:
			continue
		}
	}
}

// ReloadData replaces the current store with new data (for hot-reload).
func (e *Engine) ReloadData(data map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear existing store
	e.store = make(map[string][]map[string]interface{})

	// Reload with new data
	for key, value := range data {
		switch v := value.(type) {
		case []interface{}:
			items := make([]map[string]interface{}, 0, len(v))
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					if _, hasID := m["id"]; !hasID {
						m["id"] = uuid.New().String()
					}
					items = append(items, m)
				}
			}
			e.store[key] = items
		case map[string]interface{}:
			if _, hasID := v["id"]; !hasID {
				v["id"] = uuid.New().String()
			}
			e.store[key] = []map[string]interface{}{v}
		}
	}
}

// registerRoutes dynamically creates CRUD endpoints for each resource.
func (e *Engine) registerRoutes() {
	for resource := range e.store {
		res := resource

		e.app.Get("/"+res, e.handleGetAll(res))
		e.app.Get("/"+res+"/:id", e.handleGetByID(res))
		e.app.Post("/"+res, e.handleCreate(res))
		e.app.Put("/"+res+"/:id", e.handleUpdate(res))
		e.app.Patch("/"+res+"/:id", e.handlePatch(res))
		e.app.Delete("/"+res+"/:id", e.handleDelete(res))
	}

	// Health check
	e.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"resources": e.listResources(),
		})
	})

	// Database endpoint - returns all data
	e.app.Get("/db", func(c *fiber.Ctx) error {
		e.mu.RLock()
		defer e.mu.RUnlock()
		return c.JSON(e.store)
	})
}

// listResources returns available resource names.
func (e *Engine) listResources() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	resources := make([]string, 0, len(e.store))
	for key := range e.store {
		resources = append(resources, key)
	}
	return resources
}

// handleGetAll returns a handler with query parameter support.
// Supports: _page, _limit, _sort, _order, q (search)
func (e *Engine) handleGetAll(resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		e.mu.RLock()
		items := make([]map[string]interface{}, len(e.store[resource]))
		copy(items, e.store[resource])
		e.mu.RUnlock()

		// Full-text search: ?q=keyword
		if q := c.Query("q"); q != "" {
			q = strings.ToLower(q)
			filtered := make([]map[string]interface{}, 0)
			for _, item := range items {
				for _, v := range item {
					if strings.Contains(strings.ToLower(fmt.Sprintf("%v", v)), q) {
						filtered = append(filtered, item)
						break
					}
				}
			}
			items = filtered
		}

		// Field filters: ?field=value
		for key, values := range c.Queries() {
			if strings.HasPrefix(key, "_") || key == "q" {
				continue // Skip special params
			}
			if len(values) > 0 {
				filtered := make([]map[string]interface{}, 0)
				for _, item := range items {
					if v, ok := item[key]; ok {
						if fmt.Sprintf("%v", v) == values {
							filtered = append(filtered, item)
						}
					}
				}
				items = filtered
			}
		}

		// Sort: ?_sort=field&_order=asc|desc
		if sortField := c.Query("_sort"); sortField != "" {
			order := c.Query("_order", "asc")
			sort.Slice(items, func(i, j int) bool {
				vi := fmt.Sprintf("%v", items[i][sortField])
				vj := fmt.Sprintf("%v", items[j][sortField])
				if order == "desc" {
					return vi > vj
				}
				return vi < vj
			})
		}

		// Pagination: ?_page=1&_limit=10
		page, _ := strconv.Atoi(c.Query("_page", "0"))
		limit, _ := strconv.Atoi(c.Query("_limit", "0"))

		totalItems := len(items)

		if limit > 0 {
			start := 0
			if page > 0 {
				start = (page - 1) * limit
			}
			end := start + limit

			if start > len(items) {
				items = []map[string]interface{}{}
			} else {
				if end > len(items) {
					end = len(items)
				}
				items = items[start:end]
			}

			// Add pagination headers
			c.Set("X-Total-Count", strconv.Itoa(totalItems))
			c.Set("X-Page", strconv.Itoa(page))
			c.Set("X-Limit", strconv.Itoa(limit))
		}

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
			"error":   "not_found",
			"message": fmt.Sprintf("%s with id '%s' not found", resource, id),
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

		if _, hasID := body["id"]; !hasID {
			body["id"] = uuid.New().String()
		}

		e.mu.Lock()
		e.store[resource] = append(e.store[resource], body)
		e.mu.Unlock()

		return c.Status(fiber.StatusCreated).JSON(body)
	}
}

// handleUpdate returns a handler that replaces an existing item (PUT).
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
				body["id"] = itemID
				e.store[resource][i] = body
				return c.JSON(body)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": fmt.Sprintf("%s with id '%s' not found", resource, id),
		})
	}
}

// handlePatch returns a handler that partially updates an existing item.
func (e *Engine) handlePatch(resource string) fiber.Handler {
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
				// Merge: update only provided fields
				for k, v := range body {
					if k != "id" {
						item[k] = v
					}
				}
				e.store[resource][i] = item
				return c.JSON(item)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": fmt.Sprintf("%s with id '%s' not found", resource, id),
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
				e.store[resource] = append(items[:i], items[i+1:]...)
				return c.Status(fiber.StatusNoContent).Send(nil)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": fmt.Sprintf("%s with id '%s' not found", resource, id),
		})
	}
}

// Start runs the Fiber server.
func (e *Engine) Start(addr string) error {
	return e.app.Listen(addr)
}

// Shutdown gracefully stops the server.
func (e *Engine) Shutdown() error {
	return e.app.Shutdown()
}

// App returns the underlying Fiber app.
func (e *Engine) App() *fiber.App {
	return e.app
}

// GetStore returns a copy of the current store (for debugging/TUI).
func (e *Engine) GetStore() map[string][]map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	copy := make(map[string][]map[string]interface{})
	for k, v := range e.store {
		copy[k] = v
	}
	return copy
}
