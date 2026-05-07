# userapi

A lightweight REST API for user management built with Go and SQLite. No external database server required — SQLite creates a single `users.db` file on disk.

---

## Tech Stack

| Component | Tool | Why |
|---|---|---|
| Language | Go 1.22+ | Built-in HTTP routing with `{id}` path params |
| Database | SQLite via `modernc.org/sqlite` | Pure Go, no install needed |
| Passwords | `golang.org/x/crypto/bcrypt` | Secure hashing |
| HTTP | stdlib `net/http` | No framework needed |

---

## Project Structure

```
userapi/
├── main.go           — server setup and route registration
├── go.mod            — Go module and dependencies
├── test.sh           — shell script to test all endpoints
├── model/
│   └── user.go       — User struct and request/response types
├── db/
│   └── db.go         — SQLite connection, migrations, all queries
└── handler/
    └── user.go       — HTTP handlers for each endpoint
```

---

## Prerequisites

Install these via Homebrew:

```sh
brew install go
```

Verify:

```sh
go version   # should show go1.22 or later
```

---

## Setup

### 1. Clone or create the project directory

```sh
mkdir ~/userapi && cd ~/userapi
```

### 2. Initialise the Go module

```sh
go mod init userapi
```

This creates `go.mod` which tracks the module name and dependencies.

### 3. Install dependencies

```sh
go get modernc.org/sqlite        # SQLite driver (pure Go, no CGO)
go get golang.org/x/crypto/bcrypt # password hashing
```

### 4. Run the server

```sh
go run .
```

The server starts on `http://localhost:8080`. SQLite creates `users.db` automatically on first run.

---

## How It Works

### Database (`db/db.go`)

`db.Open()` opens the SQLite file and runs `migrate()` which creates the `users` table if it doesn't exist:

```sql
CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL,
    email      TEXT    NOT NULL UNIQUE,
    password   TEXT    NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
)
```

Timestamps are stored as Unix integers (seconds since epoch) and converted to `time.Time` when read.

`SetMaxOpenConns(1)` is set because SQLite does not support concurrent writes — this queues writes one at a time.

### Model (`model/user.go`)

Defines the `User` struct with JSON tags. The `password` field uses `json:"-"` which tells Go to never include it in JSON responses — passwords are never returned to the caller.

### Handlers (`handler/user.go`)

Each handler follows the same pattern:
1. Parse the request (path param or JSON body)
2. Validate required fields
3. Call the DB layer
4. Return JSON response

Passwords are hashed with bcrypt before storing:

```go
hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

On login, bcrypt compares the plain password against the stored hash:

```go
bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
```

### Routes (`main.go`)

Go 1.22 added method+path routing directly in stdlib:

```go
mux.HandleFunc("GET /users",        h.ListUsers)
mux.HandleFunc("POST /users",       h.CreateUser)
mux.HandleFunc("GET /users/{id}",   h.GetUser)
mux.HandleFunc("PUT /users/{id}",   h.UpdateUser)
mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
mux.HandleFunc("POST /login",       h.Login)
```

`r.PathValue("id")` extracts `{id}` from the URL.

---

## API Reference

### Create User
```sh
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'
```

### List Users
```sh
curl http://localhost:8080/users
```

### Get User by ID
```sh
curl http://localhost:8080/users/1
```

### Update User
```sh
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice2@example.com"}'
```

### Delete User
```sh
curl -X DELETE http://localhost:8080/users/1
```

### Login
```sh
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

---

## Test Script

`test.sh` creates 3 users, lists them, then deletes each one and confirms the list is empty.

```sh
# make sure the server is running first
go run .

# in another terminal
./test.sh
```

---

## Reading the Database Directly

```sh
sqlite3 users.db

# inside sqlite3
.headers on
.mode column
SELECT id, name, email FROM users;
.quit
```

One-liner:

```sh
sqlite3 -header -column users.db "SELECT id, name, email FROM users;"
```
