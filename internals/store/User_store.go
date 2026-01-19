package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var AnonymousUser = &User{}

func (u *User) IsAnonymousUser() bool {
	return u == AnonymousUser
}

type PostgresUserStore struct {
	DB *sql.DB
}

type UserStore interface {
	CreateUser(*User) error
	IsUniqueUsernameOrEmail(string, string) error
	GetUserByUserNameOrEmail(value string) (*User, error)
	GetUserToken(scope, tokenPlainText string) (*User, error)
	GetUserById(userId uuid.UUID) (*User, error)
}

func NewUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{DB: db}
}

func (pg *PostgresUserStore) CreateUser(user *User) error {
	query := `
	INSERT INTO users (username,email,password_hash,scope) 
	VALUES ($1,$2,$3,$4) 
	RETURNING id,created_at,updated_at
`
	err := pg.DB.QueryRow(query, user.UserName, user.Email, user.Password, user.Scope).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
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
		SELECT id, username, email, password_hash, scope, created_at, updated_at
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
		&user.Scope,
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

func (pg *PostgresUserStore) GetUserToken(scope, tokenPlainText string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	tokenHashHex := hex.EncodeToString(tokenHash[:])
	query := `
	 SELECT u.id, u.username, u.email, u.password_hash, u.scope, u.created_at, u.updated_at FROM users u
	 INNER JOIN tokens t on t.user_id = u.id
	 WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3
	`
	user := &User{}
	err := pg.DB.QueryRow(query, tokenHashHex, scope, time.Now()).Scan(&user.ID, &user.UserName, &user.Email, &user.Password, &user.Scope, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (pg *PostgresUserStore) GetUserById(userId uuid.UUID) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.scope,
		       u.created_at, u.updated_at
		FROM users u
		WHERE u.id = $1
	`

	user := &User{}
	err := pg.DB.QueryRow(query, userId).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.Scope,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
