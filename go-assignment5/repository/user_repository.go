package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)
import "github.com/yertaypert/go-assignment5/models"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPaginatedUsers(page int, pageSize int,
	filters map[string]string, orderBy string) (models.PaginatedResponse, error) {

	var users []models.User

	offset := (page - 1) * pageSize

	baseQuery := `
		SELECT id, name, email, gender, birth_date
		FROM users
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Dynamic filters
	for field, value := range filters {

		switch field {

		case "id":
			conditions = append(conditions, fmt.Sprintf("id = $%d", argIndex))

		case "name":
			conditions = append(conditions, fmt.Sprintf("name = $%d", argIndex))

		case "email":
			conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
			value = "%" + value + "%"

		case "gender":
			conditions = append(conditions, fmt.Sprintf("gender = $%d", argIndex))

		case "birth_date":
			conditions = append(conditions, fmt.Sprintf("birth_date = $%d", argIndex))
		}

		args = append(args, value)
		argIndex++
	}

	// Apply filters
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Default sorting
	if orderBy == "" {
		orderBy = "id"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s", orderBy)

	// Pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return models.PaginatedResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {

		var u models.User

		if err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.Gender,
			&u.BirthDate,
		); err != nil {
			return models.PaginatedResponse{}, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return models.PaginatedResponse{}, err
	}

	// Count query
	var totalCount int

	countQuery := "SELECT COUNT(*) FROM users"

	err = r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	return models.PaginatedResponse{
		Data:       users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (r *Repository) GetCommonFriends(user1 uuid.UUID, user2 uuid.UUID) ([]models.User, error) {

	var users []models.User

	query := `
	SELECT u.id, u.name, u.email, u.gender, u.birth_date
	FROM users u
	JOIN user_friends f1 
	    ON (u.id = f1.friend_id OR u.id = f1.user_id)
	JOIN user_friends f2 
	    ON (u.id = f2.friend_id OR u.id = f2.user_id)
	WHERE 
	    (f1.user_id = $1 OR f1.friend_id = $1)
	AND 
	    (f2.user_id = $2 OR f2.friend_id = $2)
	AND u.id NOT IN ($1, $2)
	`

	rows, err := r.db.Query(query, user1, user2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var u models.User

		err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.Gender,
			&u.BirthDate,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
