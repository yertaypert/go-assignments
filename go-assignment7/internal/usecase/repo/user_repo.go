package repo

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/yertaypert/go-assignment7/internal/entity"
	"github.com/yertaypert/go-assignment7/pkg"
	"github.com/yertaypert/go-assignment7/utils"
)

type UserRepo struct {
	PG *pkg.Postgres
}

func NewUserRepo(pg *pkg.Postgres) *UserRepo {
	return &UserRepo{PG: pg}
}

func (u *UserRepo) EnsureAdminUser(username, email, password string) (*entity.User, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)

	if email == "" || password == "" {
		return nil, nil
	}
	if username == "" {
		username = "admin"
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}

	const query = `
INSERT INTO users (id, username, email, password, role, verified)
VALUES ($1, $2, $3, $4, 'admin', true)
ON CONFLICT (email) DO UPDATE
SET username = EXCLUDED.username,
    password = EXCLUDED.password,
    role = 'admin',
    verified = true
RETURNING id, username, email, password, role, verified`

	var adminUser entity.User
	var id string
	err = u.PG.Conn.QueryRow(
		query,
		uuid.New(),
		username,
		email,
		hashedPassword,
	).Scan(
		&id,
		&adminUser.Username,
		&adminUser.Email,
		&adminUser.Password,
		&adminUser.Role,
		&adminUser.Verified,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure admin user: %w", err)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse admin user id: %w", err)
	}
	adminUser.ID = parsedID

	return &adminUser, nil
}

func (u *UserRepo) PromoteUser(userID string) (*entity.User, error) {
	const query = `
UPDATE users
SET role = 'admin'
WHERE id = $1
RETURNING id, username, email, password, role, verified`

	var promotedUser entity.User
	var id string
	err := u.PG.Conn.QueryRow(query, strings.TrimSpace(userID)).Scan(
		&id,
		&promotedUser.Username,
		&promotedUser.Email,
		&promotedUser.Password,
		&promotedUser.Role,
		&promotedUser.Verified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("update user role: %w", err)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	promotedUser.ID = parsedID

	return &promotedUser, nil
}

func (u *UserRepo) GetUserByID(userID string) (*entity.User, error) {
	const query = `
SELECT id, username, email, password, role, verified
FROM users
WHERE id = $1
LIMIT 1`

	var userFromDB entity.User
	var id string
	err := u.PG.Conn.QueryRow(query, strings.TrimSpace(userID)).Scan(
		&id,
		&userFromDB.Username,
		&userFromDB.Email,
		&userFromDB.Password,
		&userFromDB.Role,
		&userFromDB.Verified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	userFromDB.ID = parsedID

	return &userFromDB, nil
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	email := strings.ToLower(strings.TrimSpace(user.Email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	if err := u.PG.Conn.QueryRow(checkQuery, email).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", user.Email)
	}

	created := *user
	created.ID = uuid.New()
	created.Email = email
	if created.Role == "" {
		created.Role = "user"
	}
	created.Verified = false

	const insertQuery = `
INSERT INTO users (id, username, email, password, role, verified)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, password, role, verified`

	row := u.PG.Conn.QueryRow(
		insertQuery,
		created.ID,
		created.Username,
		created.Email,
		created.Password,
		created.Role,
		created.Verified,
	)

	var stored entity.User
	var id string
	if err := row.Scan(
		&id,
		&stored.Username,
		&stored.Email,
		&stored.Password,
		&stored.Role,
		&stored.Verified,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user insert returned no rows")
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse returned id: %w", err)
	}
	stored.ID = parsedID

	return &stored, nil
}

func (u *UserRepo) LoginUser(user *entity.LoginUserDTO) (*entity.User,
	error) {
	username := strings.TrimSpace(user.Username)
	if username == "" {
		return nil, fmt.Errorf("username is required")
	}

	const query = `
SELECT id, username, email, password, role, verified
FROM users
WHERE username = $1
LIMIT 1`

	var userFromDB entity.User
	var id string
	err := u.PG.Conn.QueryRow(query, username).Scan(
		&id,
		&userFromDB.Username,
		&userFromDB.Email,
		&userFromDB.Password,
		&userFromDB.Role,
		&userFromDB.Verified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("username not found")
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	userFromDB.ID = parsedID

	return &userFromDB, nil
}
