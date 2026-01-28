# 🚀 Insta-Mock

> **"Tu Backend listo en lo que tardas en parpadear."**

Generate instant REST APIs from JSON files with zero configuration.

## ✨ Features

- 🔥 **Instant CRUD** - GET, POST, PUT, DELETE endpoints auto-generated
- 📁 **JSON-Powered** - Your data file becomes your API
- 🔒 **Thread-Safe** - Concurrent request handling built-in
- 🌐 **CORS Enabled** - Ready for frontend integration
- 🆔 **Auto UUIDs** - IDs generated automatically for new items

---

## 📦 Installation

```bash
go install github.com/MiguelVivar/insta-mock/cmd/imock@latest
```

Or clone and build:

```bash
git clone https://github.com/MiguelVivar/insta-mock.git
cd insta-mock
go build -o imock ./cmd/imock
```

---

## 🚀 Quick Start

### 1. Create your data file

```json
// db.json
{
  "users": [{ "id": "1", "name": "Miguel", "email": "miguel@example.com" }],
  "posts": [{ "id": "1", "title": "Hello World", "authorId": "1" }]
}
```

### 2. Start the server

```bash
imock serve db.json --port 3000
```

### 3. Use your API!

| Method   | Endpoint   | Description     |
| -------- | ---------- | --------------- |
| `GET`    | `/users`   | List all users  |
| `GET`    | `/users/1` | Get user by ID  |
| `POST`   | `/users`   | Create new user |
| `PUT`    | `/users/1` | Update user     |
| `DELETE` | `/users/1` | Delete user     |
| `GET`    | `/health`  | Health check    |

---

## 📖 API Examples

### List all items

```bash
curl http://localhost:3000/users
```

### Get single item

```bash
curl http://localhost:3000/users/1
```

### Create item

```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Ana", "email": "ana@example.com"}'
```

### Update item

```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Miguel Vivar", "email": "miguel@updated.com"}'
```

### Delete item

```bash
curl -X DELETE http://localhost:3000/users/1
```

---

## 🧑‍💻 Programmatic Usage

```go
package main

import (
    "encoding/json"
    "os"
    "github.com/MiguelVivar/insta-mock/internal/server"
)

func main() {
    // Load JSON data
    file, _ := os.ReadFile("db.json")
    var data map[string]interface{}
    json.Unmarshal(file, &data)

    // Create and start engine
    engine := server.NewEngine(data)
    engine.Start(":3000")
}
```

---

## 📋 JSON Structure Rules

| Structure             | Result                        |
| --------------------- | ----------------------------- |
| `{"users": [...]}`    | Creates `/users` endpoints    |
| `{"products": [...]}` | Creates `/products` endpoints |
| Items without `id`    | UUID auto-generated           |

Each top-level key becomes a REST resource with full CRUD support.

---

## 🛠 Tech Stack

- **Server**: [Fiber v2](https://gofiber.io/) - Express-like performance
- **CLI**: [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper)
- **TUI**: [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)

---

## 📄 License

MIT © Miguel Vivar
