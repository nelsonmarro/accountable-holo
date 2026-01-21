package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nelsonmarro/verith/internal/application/service/mocks"
	"github.com/nelsonmarro/verith/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_GetAdminUsers(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	expectedAdmins := []domain.User{
		{Username: "admin1", Role: domain.RoleAdmin},
		{Username: "admin2", Role: domain.RoleAdmin},
	}

	mockRepo.On("GetUsersByRole", ctx, domain.RoleAdmin).Return(expectedAdmins, nil)

	admins, err := service.GetAdminUsers(ctx)

	assert.NoError(t, err)
	assert.Len(t, admins, 2)
	assert.Equal(t, "admin1", admins[0].Username)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		existingUser := &domain.User{
			BaseEntity:   domain.BaseEntity{ID: 1},
			Username:     "admin",
			PasswordHash: "oldhash",
			Role:         domain.RoleAdmin,
		}

		mockRepo.On("GetUserByUsername", ctx, "admin").Return(existingUser, nil).Once()
		
		// Verificamos que UpdateUser sea llamado con una contraseña hasheada diferente a la original
		mockRepo.On("UpdateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
			return u.Username == "admin" && u.PasswordHash != "oldhash"
		})).Return(nil).Once()

		err := service.ResetPassword(ctx, "admin", "NewSecretPass123!")

		assert.NoError(t, err)
		
		// Verificar que la nueva contraseña es válida bcrypt
		// Nota: En un test real unitario, no podemos verificar el valor exacto del hash generado dentro del método 
		// sin exponerlo, pero el MatchedBy ya valida que cambió.
		mockRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo.On("GetUserByUsername", ctx, "ghost").Return((*domain.User)(nil), errors.New("user not found")).Once()

		err := service.ResetPassword(ctx, "ghost", "password")

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		mockRepo.AssertNotCalled(t, "UpdateUser")
	})
	
	t.Run("Verify Hash Valid", func(t *testing.T) {
		// Este test asegura que ResetPassword realmente hashea la contraseña
		user := &domain.User{Username: "admin"}
		var capturedUser *domain.User
		
		mockRepo.On("GetUserByUsername", ctx, "admin").Return(user, nil).Once()
		mockRepo.On("UpdateUser", ctx, mock.MatchedBy(func(u *domain.User) bool {
			capturedUser = u
			return true
		})).Return(nil).Once()

		password := "SecurePass123"
		_ = service.ResetPassword(ctx, "admin", password)

		assert.NotNil(t, capturedUser)
		assert.NotEqual(t, password, capturedUser.PasswordHash) // No debe ser texto plano
		
		err := bcrypt.CompareHashAndPassword([]byte(capturedUser.PasswordHash), []byte(password))
		assert.NoError(t, err, "Password hash should be valid for the input password")
	})
}
