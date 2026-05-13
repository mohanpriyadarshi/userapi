package db

import (
	"database/sql"
	"time"
	"userapi/model"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// Open opens a SQLite database at the provided path, limits connections to one
// to avoid SQLite write concurrency issues, and runs the schema migration.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes
	d := &DB{sqlDB}
	return d, d.migrate()
}

// migrate creates the users table if it does not already exist.
func (d *DB) migrate() error {
	_, err := d.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL,
			email      TEXT    NOT NULL UNIQUE,
			password   TEXT    NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (unixepoch()),
			updated_at INTEGER NOT NULL DEFAULT (unixepoch())
		)
	`)
	return err
}

// scanUser scans database row fields into a model.User and converts Unix timestamps
// into time.Time values.
func scanUser(row interface{ Scan(...any) error }) (*model.User, error) {
	var u model.User
	var createdAt, updatedAt int64
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)
	return &u, nil
}

const selectUser = `SELECT id, name, email, created_at, updated_at FROM users`

// ListUsers returns a paginated slice of users using the provided limit and offset.
// This avoids loading the entire users table into memory for large datasets.
func (d *DB) ListUsers(limit, offset int) ([]model.User, error) {
	query := selectUser + " LIMIT ? OFFSET ?"
	rows, err := d.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, nil
}

// GetUserByID fetches a single user by their numeric ID.
func (d *DB) GetUserByID(id int64) (*model.User, error) {
	return scanUser(d.QueryRow(selectUser+` WHERE id = ?`, id))
}

// GetUserByEmail retrieves a user's password hash and public profile by email.
// This is used for authentication without exposing password data in the user model.
func (d *DB) GetUserByEmail(email string) (string, *model.User, error) {
	var u model.User
	var password string
	var createdAt, updatedAt int64
	err := d.QueryRow(
		`SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &password, &createdAt, &updatedAt)
	if err != nil {
		return "", nil, err
	}
	u.CreatedAt = time.Unix(createdAt, 0)
	u.UpdatedAt = time.Unix(updatedAt, 0)
	return password, &u, nil
}

// CreateUser inserts a new user and returns the created record with its new ID.
func (d *DB) CreateUser(name, email, password string) (*model.User, error) {
	result, err := d.Exec(
		`INSERT INTO users (name, email, password) VALUES (?, ?, ?)`,
		name, email, password,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return d.GetUserByID(id)
}

// UpdateUser modifies an existing user's name and email and returns the updated record.
func (d *DB) UpdateUser(id int64, name, email string) (*model.User, error) {
	res, err := d.Exec(
		`UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?`,
		name, email, time.Now().Unix(), id,
	)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return d.GetUserByID(id)
}

// DeleteUser removes a user by ID and returns sql.ErrNoRows if no record existed.
func (d *DB) DeleteUser(id int64) error {
	res, err := d.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
