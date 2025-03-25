package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"
	"github.com/flexer2006/case-back-restaurant-go/internal/repository"

	"go.uber.org/zap"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type UserUseCase interface {
	GetUser(ctx context.Context, id string) (*domain.User, error)

	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	CreateUser(ctx context.Context, user *domain.User) (string, error)

	UpdateUser(ctx context.Context, user *domain.User) error
}

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (u *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return u.userRepo.GetByEmail(ctx, email)
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "creating new user",
		zap.String("email", user.Email),
		zap.String("name", user.Name))

	existingUser, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		log.Warn(ctx, "attempt to create user with existing email",
			zap.String("email", user.Email))
		return "", ErrEmailExists
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := u.userRepo.Create(ctx, user); err != nil {
		log.Error(ctx, "failed to create user",
			zap.String("email", user.Email),
			zap.Error(err))
		return "", err
	}

	log.Info(ctx, "user successfully created",
		zap.String("userID", user.ID),
		zap.String("email", user.Email))
	return user.ID, nil
}

func (u *userUseCase) UpdateUser(ctx context.Context, user *domain.User) error {
	log, _ := logger.FromContext(ctx)
	log.Info(ctx, "updating user",
		zap.String("userID", user.ID),
		zap.String("email", user.Email))

	existingUser, err := u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		log.Error(ctx, "failed to get user by ID",
			zap.String("userID", user.ID),
			zap.Error(err))
		return err
	}
	if existingUser == nil {
		log.Warn(ctx, "user not found",
			zap.String("userID", user.ID))
		return ErrUserNotFound
	}

	if existingUser.Email != user.Email {
		userWithSameEmail, err := u.userRepo.GetByEmail(ctx, user.Email)
		if err == nil && userWithSameEmail != nil && userWithSameEmail.ID != user.ID {
			log.Warn(ctx, "attempt to update user with email that already exists",
				zap.String("email", user.Email),
				zap.String("userID", user.ID),
				zap.String("existingUserID", userWithSameEmail.ID))
			return ErrEmailExists
		}
	}

	user.UpdatedAt = time.Now()
	user.CreatedAt = existingUser.CreatedAt

	if err := u.userRepo.Update(ctx, user); err != nil {
		log.Error(ctx, "failed to update user",
			zap.String("userID", user.ID),
			zap.Error(err))
		return err
	}

	log.Info(ctx, "user successfully updated",
		zap.String("userID", user.ID),
		zap.String("email", user.Email))
	return nil
}
