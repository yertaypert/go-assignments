package users

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yertaypert/go-assignment3/internal/repository/_postgres"
	"github.com/yertaypert/go-assignment3/pkg/modules"
)

type Repository struct {
	db               *sqlx.DB
	executionTimeout time.Duration
}

func NewRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db.DB,
		executionTimeout: 5 * time.Second,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	query := "SELECT * FROM users"
	err := r.db.Select(&users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

var (
	ErrUserNotFound = errors.New("user not found")
)

func (r *Repository) CreateUser(user *modules.User) (int, error) {
	query := `INSERT INTO users (name, email, age, created_at) 
			  VALUES ($1, $2, $3, NOW()) 
			  RETURNING id
	`
	var id int
	err := r.db.Get(&id, query, user.Name, user.Email, user.Age)
	if err != nil {
		return 0, fmt.Errorf("error creating user: %w", err)
	}
	return id, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var user modules.User
	query := `SELECT id, name, email, age, created_at FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return &user, nil
}

func (r *Repository) UpdateUser(user *modules.User) error {
	query := `UPDATE users
			  SET name = $1, email = $2, age = $3
			  WHERE id = $4
   			  `
	result, err := r.db.Exec(query, user.Name, user.Email, user.Age, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return 0, fmt.Errorf("error deleting user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return 0, ErrUserNotFound
	}
	return rowsAffected, nil
}
