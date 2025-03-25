package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"github.com/flexer2006/case-back-restaurant-go/internal/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var ErrUserNotFound = errors.New(common.ErrUserNotFound)

type UserRepository struct {
	*Repository
}

func NewUserRepository(repository *Repository) *UserRepository {
	return &UserRepository{
		Repository: repository,
	}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	_, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrScanUser, err)
	}

	const query = `
		SELECT id, name, email, phone, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	return r.getUserByQuery(ctx, query, id)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	_, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrScanUser, err)
	}

	const query = `
		SELECT id, name, email, phone, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	return r.getUserByQuery(ctx, query, email)
}

func (r *UserRepository) getUserByQuery(ctx context.Context, query string, param string) (*domain.User, error) {
	logger, err := logger.FromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrScanUser, err)
	}

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		logger.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrGetQueryExecutor, err)
	}
	defer release()

	var user domain.User
	err = executor.QueryRow(ctx, query, param).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", common.ErrUserNotFound, err)
		}
		logger.Error(ctx, common.ErrScanUser,
			zap.String("param", param),
			zap.Error(err))
		return nil, fmt.Errorf("%s: %w", common.ErrScanUser, err)
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	log, _ := logger.FromContext(ctx)

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	const query = `
		INSERT INTO users (id, name, email, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	exists, err := r.checkEmailExists(ctx, user.Email, executor)
	if err != nil {
		log.Error(ctx, common.ErrCheckEmailExistence,
			zap.String("email", user.Email),
			zap.Error(err))
		return err
	}
	if exists {
		return errors.New(common.ErrEmailAlreadyExists)
	}

	_, err = executor.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrCreateUser,
			zap.String("email", user.Email),
			zap.Error(err))
		return err
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	log, _ := logger.FromContext(ctx)

	const query = `
		UPDATE users
		SET name = $2, email = $3, phone = $4, updated_at = $5
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	executor, release, err := r.GetExecutor(ctx)
	if err != nil {
		log.Error(ctx, common.ErrGetQueryExecutor, zap.Error(err))
		return err
	}
	defer release()

	currentUser, err := r.GetByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn(ctx, common.WarnUserForUpdateNotFound,
				zap.String("userID", user.ID))
			return ErrUserNotFound
		}
		log.Error(ctx, common.ErrGetCurrentUser,
			zap.String("userID", user.ID),
			zap.Error(err))
		return err
	}

	if currentUser.Email != user.Email {
		exists, err := r.checkEmailExists(ctx, user.Email, executor)
		if err != nil {
			log.Error(ctx, common.ErrCheckEmailExistence,
				zap.String("email", user.Email),
				zap.Error(err))
			return err
		}
		if exists {
			return errors.New(common.ErrEmailAlreadyExists)
		}
	}

	commandTag, err := executor.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.UpdatedAt,
	)
	if err != nil {
		log.Error(ctx, common.ErrUpdateUser,
			zap.String("userID", user.ID),
			zap.Error(err))
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) checkEmailExists(ctx context.Context, email string, executor DBExecutor) (bool, error) {
	const query = `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	var exists bool
	err := executor.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
