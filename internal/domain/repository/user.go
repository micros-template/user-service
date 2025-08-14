package repository

import (
	"context"
	"errors"
	"fmt"

	"10.1.20.130/dropping/sharedlib/model"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	_db "10.1.20.130/dropping/user-service/internal/infrastructure/database"
	"10.1.20.130/dropping/user-service/internal/infrastructure/logger"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type (
	UserRepository interface {
		CreateNewUser(*model.User) error
		QueryUserByUserId(string) (*model.User, error)
		UpdateUser(*model.User) error
		DeleteUser(userId string) error
	}
	userRepository struct {
		pgx        _db.Querier
		logger     zerolog.Logger
		logEmitter logger.LoggerInfra
	}
)

func NewUserRepository(pgx _db.Querier, logEmitter logger.LoggerInfra, logger zerolog.Logger) UserRepository {
	return &userRepository{
		pgx:        pgx,
		logger:     logger,
		logEmitter: logEmitter,
	}
}

func (a *userRepository) DeleteUser(userId string) error {
	// [IMRPOVE] change to soft delete
	query, args, err := sq.Delete("users").
		Where(sq.Eq{"id": userId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_BUILD_QUERY.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_BUILD_QUERY
	}

	cmdTag, err := a.pgx.Exec(context.Background(), query, args...)
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_DELETE_USER.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_DELETE_USER
	}
	if cmdTag.RowsAffected() == 0 {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", fmt.Sprintf("%s. user_id: %s", dto.Err_NOTFOUND_USER_NOT_FOUND.Error(), userId)); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_NOTFOUND_USER_NOT_FOUND
	}
	return nil
}

func (a *userRepository) UpdateUser(user *model.User) error {
	query, args, err := sq.Update("users").
		Set("full_name", user.FullName).
		Set("image", user.Image).
		Set("email", user.Email).
		Set("password", user.Password).
		Set("verified", user.Verified).
		Set("two_factor_enabled", user.TwoFactorEnabled).
		Set("updated_at", sq.Expr("CURRENT_TIMESTAMP")).
		Where(sq.Eq{"id": user.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_BUILD_QUERY.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_BUILD_QUERY
	}

	cmdTag, err := a.pgx.Exec(context.Background(), query, args...)
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_UPDATE_USER.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_UPDATE_USER
	}
	if cmdTag.RowsAffected() == 0 {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", fmt.Sprintf("%s. user_id: %s", dto.Err_NOTFOUND_USER_NOT_FOUND.Error(), user.ID)); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_NOTFOUND_USER_NOT_FOUND
	}
	return nil
}

func (a *userRepository) CreateNewUser(user *model.User) error {
	query, args, err := sq.Insert("users").
		Columns("id", "full_name", "image", "email", "password", "verified", "two_factor_enabled").
		Values(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_BUILD_QUERY.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_BUILD_QUERY
	}
	row := a.pgx.QueryRow(context.Background(), query, args...)
	if err := row.Scan(&user.ID); err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_INSERT_USER.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_INTERNAL_FAILED_INSERT_USER
	}
	return nil
}

func (a *userRepository) QueryUserByUserId(userId string) (*model.User, error) {
	var user model.User
	query, args, err := sq.Select("id", "full_name", "image", "email", "password", "verified", "two_factor_enabled").
		From("users").
		Where(sq.Eq{"id": userId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_BUILD_QUERY.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return nil, dto.Err_INTERNAL_FAILED_BUILD_QUERY
	}
	row := a.pgx.QueryRow(context.Background(), query, args...)
	err = row.Scan(&user.ID, &user.FullName, &user.Image, &user.Email, &user.Password, &user.Verified, &user.TwoFactorEnabled)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			go func() {
				if err := a.logEmitter.EmitLog("WARN", fmt.Sprintf("%s user_id: %s", dto.Err_NOTFOUND_USER_NOT_FOUND.Error(), userId)); err != nil {
					a.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			return nil, dto.Err_NOTFOUND_USER_NOT_FOUND
		}
		go func() {
			if err := a.logEmitter.EmitLog("ERR", dto.Err_INTERNAL_FAILED_SCAN_USER.Error()); err != nil {
				a.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return nil, dto.Err_INTERNAL_FAILED_SCAN_USER
	}
	return &user, nil

}
