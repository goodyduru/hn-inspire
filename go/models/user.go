package models

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
	IsAdmin  bool
	Joined   time.Time
	Karma    int
	About    string
}

var ErrNotUnique = fmt.Errorf("%s user with that username already exists%s", "A", ".")

func (u *User) Create() error {
	err := u.usernameExists()
	if err != nil {
		return err
	}
	row := db.QueryRow(`INSERT INTO users (username, password, is_admin) VALUES ($1, $2, $3) RETURNING id, date_joined`,
		u.Username, hash(u.Password), u.IsAdmin)
	return row.Scan(&u.ID, &u.Joined)
}

func (u *User) usernameExists() error {
	id := 0
	row := db.QueryRow("SELECT id from users WHERE username = $1", u.Username)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return ErrNotUnique
}

func (u *User) ReadByUsername() error {
	return db.QueryRow("SELECT id, COALESCE(email, ''), password, is_admin, date_joined, karma, COALESCE(about, '') FROM users WHERE username = $1", u.Username).
		Scan(&u.ID, &u.Email, &u.Password, &u.IsAdmin, &u.Joined, &u.Karma, &u.About)
}

func (u *User) Update() error {
	_, err := db.Exec(`UPDATE users SET (email, about) = ($2, $3) WHERE id=$1`, u.ID, u.Email, u.About)
	return err
}

func hash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}
