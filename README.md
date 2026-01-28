# 🚀 Insta-Mock

> **"Tu Backend listo en lo que tardas en parpadear."**

Generate instant REST APIs from JSON files with zero configuration.

## ✨ Features

| Feature               | Description                                  |
| --------------------- | -------------------------------------------- |
| 🔥 **Instant CRUD**   | GET, POST, PUT, PATCH, DELETE auto-generated |
| 📁 **JSON-Powered**   | Your data file becomes your API              |
| 🧠 **Smart Data Gen** | Generate fake data from field names          |
| 🔄 **Hot Reload**     | Watch file changes, auto-reload (`--watch`)  |
| 💥 **Chaos Mode**     | Simulate failures/latency (`--chaos`)        |
| 🔍 **Query Params**   | Pagination, sorting, filtering, search       |
| 🌐 **CORS Enabled**   | Ready for frontend integration               |

---

## 📦 Installation

```bash
go install github.com/MiguelVivar/insta-mock/cmd/imock@main
```

---

## 🚀 Quick Start

### 1. Create your data file

```json
{
  "users": [{ "id": "1", "name": "Miguel", "email": "miguel@example.com" }],
  "posts": [{ "id": "1", "title": "Hello World", "authorId": "1" }]
}
```

### 2. Start the server

```bash
imock serve db.json --port 3000
```

### 3. Advanced options

```bash
# Generate 10 fake items per resource
imock serve db.json --count 10

# Enable hot-reload (auto-reload on file changes)
imock serve db.json --watch

# Enable chaos mode (random failures/latency)
imock serve db.json --chaos

# Combine all features
imock serve db.json -p 8080 -c 20 -w --chaos
```

---

## 📖 API Reference

### Endpoints

| Method   | Endpoint         | Description                  |
| -------- | ---------------- | ---------------------------- |
| `GET`    | `/:resource`     | List all (with query params) |
| `GET`    | `/:resource/:id` | Get by ID                    |
| `POST`   | `/:resource`     | Create new item              |
| `PUT`    | `/:resource/:id` | Replace item                 |
| `PATCH`  | `/:resource/:id` | Partial update               |
| `DELETE` | `/:resource/:id` | Delete item                  |
| `GET`    | `/db`            | Get entire database          |
| `GET`    | `/health`        | Health check                 |

### Query Parameters

```bash
# Pagination
GET /users?_page=1&_limit=10

# Sorting
GET /users?_sort=name&_order=desc

# Full-text search
GET /users?q=miguel

# Field filtering
GET /users?role=admin
GET /posts?authorId=1
```

---

## 🧠 Smart Data Generation

Field names are analyzed to generate appropriate fake data:

| Field Pattern        | Generated Data             |
| -------------------- | -------------------------- |
| `email`, `correo`    | `ana.garcia@example.com`   |
| `name`, `nombre`     | `Miguel Rodríguez`         |
| `phone`, `telefono`  | `+1-555-123-4567`          |
| `title`, `titulo`    | `Lorem ipsum sentence`     |
| `price`, `precio`    | `$42.99`                   |
| `id`, `*_id`         | UUID auto-generated        |
| `url`, `website`     | `https://example.com/path` |
| `image`, `avatar`    | Image URL                  |
| `address`, `street`  | Street address             |
| `city`, `ciudad`     | City name                  |
| `company`, `empresa` | Company name               |

---

## 🛠 CLI Reference

```
imock serve <json-file> [flags]

Flags:
  -p, --port string   Port to run the server (default "3000")
  -c, --count int     Generate N fake items per resource
  -w, --watch         Watch file for changes (hot-reload)
      --chaos         Enable chaos mode (random failures)
  -h, --help          Help for serve
```

---

## 📄 License

MIT © Miguel Vivar
