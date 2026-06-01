package entity

import (
	"encoding/json"
	"strings"
)

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional: "user" (default) or "admin"
}

func (c *CreateUserDTO) UnmarshalJSON(data []byte) error {
	type Alias CreateUserDTO

	var aux Alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Username = strings.TrimSpace(aux.Username)
	c.Email = strings.TrimSpace(aux.Email)
	c.Password = strings.TrimSpace(aux.Password)
	c.Role = strings.TrimSpace(aux.Role)

	return nil
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
