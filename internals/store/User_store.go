package store

import (
	"database/sql"
	"time"
)

type User struct {
	ID string`json:"id"`
	UserName string `json:"username"`
	Email string `json:"email"`
	Password string `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostgresUserStore struct {
	DB *sql.DB
}

type UserStore interface {
	CreateUser(*User) error
}

func NewUserStore(db *sql.DB) (*PostgresUserStore) {
return &PostgresUserStore{DB: db}
}

func (pg *PostgresUserStore) CreateUser(user *User) error {
query := `
	INSERT INTO users (username,email,password_hash) 
	VALUES ($1,$2,$3) 
	RETURNING id,created_at,updated_at
`
err := pg.DB.QueryRow(query,user.UserName,user.Email,user.Password).Scan(&user.ID,&user.CreatedAt,&user.UpdatedAt)
if err != nil {
	return err
}
return nil
}

