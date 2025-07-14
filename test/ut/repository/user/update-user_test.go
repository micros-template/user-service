package repository_test

import (
	"fmt"
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"github.com/dropboks/sharedlib/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type UpdateUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
}

func (g *UpdateUserRepositorySuite) SetupSuite() {
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	g.NoError(err)
	g.mockPgx = pgxMock
	g.userRepository = repository.NewUserRepository(pgxMock, logger)
}

func TestUpdateUserRepositorySuite(t *testing.T) {
	suite.Run(t, &UpdateUserRepositorySuite{})
}

func (g *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_Success() {
	email := "update-success@example.com"
	image := "updated-image.png"
	user := &model.User{
		ID:               email,
		FullName:         "updated_user",
		Image:            &image,
		Email:            email,
		Password:         "updatedpassword",
		Verified:         true,
		TwoFactorEnabled: true,
	}

	updateQuery := `UPDATE users SET full_name = \$1, image = \$2, email = \$3, password = \$4, verified = \$5, two_factor_enabled = \$6, updated_at = CURRENT_TIMESTAMP WHERE id = \$7`
	g.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := g.userRepository.UpdateUser(user)
	g.NoError(err)
}

func (g *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_NotFound() {
	email := "notfound@example.com"
	user := &model.User{
		ID:               email,
		FullName:         "nonexistent_user",
		Image:            nil,
		Email:            email,
		Password:         "password",
		Verified:         false,
		TwoFactorEnabled: false,
	}

	updateQuery := `UPDATE users SET full_name = \$1, image = \$2, email = \$3, password = \$4, verified = \$5, two_factor_enabled = \$6, updated_at = CURRENT_TIMESTAMP WHERE id = \$7`
	g.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := g.userRepository.UpdateUser(user)
	g.ErrorIs(err, dto.Err_NOTFOUND_USER_NOT_FOUND)
}

func (g *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_QueryError() {
	email := "query-error@example.com"
	user := &model.User{
		ID:               email,
		FullName:         "error_user",
		Image:            nil,
		Email:            email,
		Password:         "password",
		Verified:         false,
		TwoFactorEnabled: false,
	}

	updateQuery := `UPDATE users SET full_name = \$1, image = \$2, email = \$3, password = \$4, verified = \$5, two_factor_enabled = \$6, updated_at = CURRENT_TIMESTAMP WHERE id = \$7`
	g.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnError(fmt.Errorf("query execution failed"))

	err := g.userRepository.UpdateUser(user)
	g.ErrorIs(err, dto.Err_INTERNAL_FAILED_UPDATE_USER)
}
