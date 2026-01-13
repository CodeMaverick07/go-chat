package store

import (
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostgresUserStore struct {
	DB *sql.DB
}

type UserStore interface {
	CreateUser(*User) error
	IsUniqueUsernameOrEmail(string, string) error
	GetUserByUserNameOrEmail(value string) (*User, error)
}

func NewUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{DB: db}
}

func (pg *PostgresUserStore) CreateUser(user *User) error {
	query := `
	INSERT INTO users (username,email,password_hash) 
	VALUES ($1,$2,$3) 
	RETURNING id,created_at,updated_at
`
	err := pg.DB.QueryRow(query, user.UserName, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgresUserStore) IsUniqueUsernameOrEmail(
	value string,
	what string,
) error {

	var exists bool
	var query string

	switch what {
	case "email":
		query = `
			SELECT EXISTS (
				SELECT 1 FROM users WHERE email = $1
			)
		`
	case "username":
		query = `
			SELECT EXISTS (
				SELECT 1 FROM users WHERE username = $1
			)
		`
	default:
		return errors.New("invalid uniqueness check type")
	}

	err := pg.DB.QueryRow(query, value).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(what + " already exists")
	}

	return nil
}

func (pg *PostgresUserStore) GetUserByUserNameOrEmail(value string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1 OR username = $1
		LIMIT 1
	`

	user := &User{}

	err := pg.DB.QueryRow(query, value).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}
