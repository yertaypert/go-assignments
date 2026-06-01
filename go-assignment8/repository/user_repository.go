package repository

//go:generate mockgen -source=user_repository.go -destination=mock_user_repository.go -package=repository

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetUserByID(id int) (*User, error)
	CreateUser(user *User) error
	GetByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}
