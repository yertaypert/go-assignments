package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yertaypert/go-assignment8/repository"
	"go.uber.org/mock/gomock"
)

func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)
	result, err := userService.GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)
	err := userService.CreateUser(user)
	assert.NoError(t, err)
}

func TestGetByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().GetByEmail("test@example.com").Return(user, nil)
	result, err := userService.GetByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestUpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockUserRepository(ctrl)
	userService := NewUserService(mockRepo)
	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().UpdateUser(user).Return(nil)
	err := userService.UpdateUser(user)
	assert.NoError(t, err)
}

func TestRegisterUser(t *testing.T) {
	t.Run("User already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)
		email := "existing@example.com"
		user := &repository.User{Name: "Existing User"}

		mockRepo.EXPECT().GetByEmail(email).Return(&repository.User{ID: 1, Name: "Existing User"}, nil)
		err := userService.RegisterUser(user, email)
		assert.Error(t, err)
		assert.Equal(t, "user with this email already exists", err.Error())
	})

	t.Run("New User -> Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)
		email := "new@example.com"
		user := &repository.User{Name: "New User"}

		mockRepo.EXPECT().GetByEmail(email).Return(nil, nil)
		mockRepo.EXPECT().CreateUser(user).Return(nil)
		err := userService.RegisterUser(user, email)
		assert.NoError(t, err)
	})

	t.Run("Repository error on CreateUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)
		email := "new@example.com"
		user := &repository.User{Name: "New User"}

		mockRepo.EXPECT().GetByEmail(email).Return(nil, nil)
		mockRepo.EXPECT().CreateUser(user).Return(errors.New("db error"))
		err := userService.RegisterUser(user, email)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})
}

func TestUpdateUserName(t *testing.T) {
	t.Run("Empty name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		err := userService.UpdateUserName(1, "")
		assert.Error(t, err)
		assert.Equal(t, "name cannot be empty", err.Error())
	})

	t.Run("User not found / repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		mockRepo.EXPECT().GetUserByID(1).Return(nil, errors.New("user not found"))
		err := userService.UpdateUserName(1, "New Name")
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})

	t.Run("Successful update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)
		id := 1
		newName := "Updated Name"
		user := &repository.User{ID: id, Name: "Old Name"}

		mockRepo.EXPECT().GetUserByID(id).Return(user, nil)

		mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(u *repository.User) error {
			if u.Name != newName {
				t.Errorf("Expected user name to be %s, got %s", newName, u.Name)
			}
			return nil
		})

		err := userService.UpdateUserName(id, newName)
		assert.NoError(t, err)
	})

	t.Run("UpdateUser Fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)
		id := 1
		newName := "Updated Name"
		user := &repository.User{ID: id, Name: "Old Name"}

		mockRepo.EXPECT().GetUserByID(id).Return(user, nil)
		mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(errors.New("update error"))

		err := userService.UpdateUserName(id, newName)
		assert.Error(t, err)
		assert.Equal(t, "update error", err.Error())
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("Attempt to delete admin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		// Admin is ID 1
		err := userService.DeleteUser(1)
		assert.Error(t, err)
		assert.Equal(t, "it is not allowed to delete admin user", err.Error())
	})

	t.Run("Successful delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		// Verify user was deleted
		mockRepo.EXPECT().DeleteUser(2).Return(nil)
		err := userService.DeleteUser(2)
		assert.NoError(t, err)
	})

	t.Run("Repository Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		mockRepo.EXPECT().DeleteUser(2).Return(errors.New("db error"))
		err := userService.DeleteUser(2)
		assert.Error(t, err)
		assert.Equal(t, "db error", err.Error())
	})
}
