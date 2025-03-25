package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/pkg/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

var _ = newTestContext

func createTestUser() *domain.User {
	return &domain.User{
		ID:        uuid.New().String(),
		Name:      "test user",
		Email:     "test@example.com",
		Phone:     "+7 (123) 456-78-90",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestUserUseCase_GetUser(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	userID := uuid.New().String()
	expectedUser := createTestUser()
	expectedUser.ID = userID

	mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	result, err := useCase.GetUser(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserNotFound(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	userID := uuid.New().String()
	expectedError := errors.New("user not found")

	mockUserRepo.On("GetByID", ctx, userID).Return(nil, expectedError)

	result, err := useCase.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserByEmail(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	email := "test@example.com"
	expectedUser := createTestUser()
	expectedUser.Email = email

	mockUserRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil)

	result, err := useCase.GetUserByEmail(ctx, email)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_GetUserByEmailNotFound(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	email := "nonexistent@example.com"
	expectedError := errors.New("user not found")

	mockUserRepo.On("GetByEmail", ctx, email).Return(nil, expectedError)

	result, err := useCase.GetUserByEmail(ctx, email)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_CreateUser(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	newUser := &domain.User{
		Name:  "new user",
		Email: "new@example.com",
		Phone: "+7 (987) 654-32-10",
	}

	expectedID := "test-user-id"

	mockUserRepo.On("GetByEmail", ctx, newUser.Email).Return(nil, errors.New("user not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = expectedID
	}).Return(nil)

	id, err := useCase.CreateUser(ctx, newUser)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, expectedID, id)
	assert.Equal(t, expectedID, newUser.ID)
	assert.False(t, newUser.CreatedAt.IsZero())
	assert.False(t, newUser.UpdatedAt.IsZero())
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_CreateUserEmailExists(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	existingUser := createTestUser()
	newUser := &domain.User{
		Name:  "duplicate user",
		Email: existingUser.Email,
		Phone: "+7 (987) 654-32-10",
	}

	mockUserRepo.On("GetByEmail", ctx, newUser.Email).Return(existingUser, nil)

	id, err := useCase.CreateUser(ctx, newUser)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrEmailExists, err)
	assert.Empty(t, id)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_UpdateUser(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	existingUser := createTestUser()
	updatedUser := &domain.User{
		ID:        existingUser.ID,
		Name:      "updated name",
		Email:     existingUser.Email,
		Phone:     "+7 (987) 654-32-10",
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}

	oldUpdateTime := updatedUser.UpdatedAt

	time.Sleep(1 * time.Millisecond)

	mockUserRepo.On("GetByID", ctx, updatedUser.ID).Return(existingUser, nil)
	mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	err := useCase.UpdateUser(ctx, updatedUser)

	assert.NoError(t, err)
	assert.True(t, updatedUser.UpdatedAt.After(oldUpdateTime))
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_UpdateUserNotFound(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	updatedUser := createTestUser()

	mockUserRepo.On("GetByID", ctx, updatedUser.ID).Return(nil, usecase.ErrUserNotFound)

	err := useCase.UpdateUser(ctx, updatedUser)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrUserNotFound, err)
	mockUserRepo.AssertExpectations(t)
}

func TestUserUseCase_UpdateUserEmailExists(t *testing.T) {
	ctx := newTestContext()
	mockUserRepo := new(MockUserRepository)

	useCase := usecase.NewUserUseCase(mockUserRepo)

	existingUser := createTestUser()
	anotherUser := createTestUser()
	anotherUser.ID = uuid.New().String()
	anotherUser.Email = "another@example.com"

	updatedUser := &domain.User{
		ID:        existingUser.ID,
		Name:      existingUser.Name,
		Email:     "another@example.com",
		Phone:     existingUser.Phone,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
	}

	mockUserRepo.On("GetByID", ctx, updatedUser.ID).Return(existingUser, nil)
	mockUserRepo.On("GetByEmail", ctx, updatedUser.Email).Return(anotherUser, nil)

	err := useCase.UpdateUser(ctx, updatedUser)

	assert.Error(t, err)
	assert.Equal(t, usecase.ErrEmailExists, err)
	mockUserRepo.AssertExpectations(t)
}
