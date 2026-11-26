# Simple REST API (Go, no database)

A minimal CRUD API written with the Go standard library only. Data lives in
memory, so everything resets when the server restarts.

## Run

```bash
go run .
# or
go build -o simple-api . && ./simple-api
```

The server listens on `:8080` by default. Override with `PORT`:

```bash
PORT=3000 go run .
```

## Endpoints

| Method | Path          | Description        |
| ------ | ------------- | ------------------ |
| GET    | `/`           | Route index        |
| GET    | `/health`     | Health check       |
| GET    | `/items`      | List all items     |
| POST   | `/items`      | Create an item     |
| GET    | `/items/{id}` | Get one item       |
| PUT    | `/items/{id}` | Update an item     |
| DELETE | `/items/{id}` | Delete an item     |

## Examples

```bash
# Create
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"Learn Go","description":"Build a small API","done":false}'

# List
curl http://localhost:8080/items

# Get
curl http://localhost:8080/items/1

# Update
curl -X PUT http://localhost:8080/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Learn Go","description":"Build a small API","done":true}'

# Delete
curl -X DELETE http://localhost:8080/items/1
```

Item JSON:

```json
{
  "id": "1",
  "name": "Learn Go",
  "description": "Build a small API",
  "done": false,
  "createdAt": "2026-11-26T08:00:00Z",
  "updatedAt": "2026-11-26T08:00:00Z"
}
```

Errors use a consistent shape:

```json
{ "error": "item not found" }
```
