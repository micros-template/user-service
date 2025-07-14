package repository_test

import (
	"fmt"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"github.com/dropboks/sharedlib/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type GetUserByEmailRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
}

func (g *GetUserByEmailRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	g.NoError(err)
	g.mockPgx = pgxMock
	g.userRepository = repository.NewUserRepository(pgxMock, logger)
}

func TestGetUserByEmailRepositorySuite(t *testing.T) {
	suite.Run(t, &GetUserByEmailRepositorySuite{})
}

func (g *GetUserByEmailRepositorySuite) TestUserRepository_GetUserByEmail_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	image := "image.png"
	expectedUser := &model.User{
		ID:               email,
		FullName:         "test_user",
		Image:            &image,
		Email:            email,
		Password:         "hashedpassword",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	rows := pgxmock.NewRows([]string{
		"id", "full_name", "image", "email", "password", "verified", "two_factor_enabled",
	}).AddRow(
		expectedUser.ID,
		expectedUser.FullName,
		expectedUser.Image,
		expectedUser.Email,
		expectedUser.Password,
		expectedUser.Verified,
		expectedUser.TwoFactorEnabled,
	)

	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE email = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)

	user, err := g.userRepository.QueryUserByEmail(email)
	g.NoError(err)
	g.Equal(expectedUser, user)
}

func (g *GetUserByEmailRepositorySuite) TestUserRepository_GetUserByEmail_NotFound() {
	email := "notfound@example.com"
	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE email = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(email).WillReturnError(pgx.ErrNoRows)

	user, err := g.userRepository.QueryUserByEmail(email)
	g.Nil(user)
	g.ErrorIs(err, dto.Err_NOTFOUND_USER_NOT_FOUND)
}

func (g *GetUserByEmailRepositorySuite) TestUserRepository_GetUserByEmail_ScanError() {
	email := "scanerror@example.com"
	rows := pgxmock.NewRows([]string{
		"id", "full_name", "image", "email", "password", "verified", "two_factor_enabled",
	}).AddRow(
		1, // should be string, but int to cause scan error
		"Test User",
		"image.png",
		email,
		"hashedpassword",
		true,
		false,
	)
	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE email = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)

	user, err := g.userRepository.QueryUserByEmail(email)
	g.Nil(user)
	g.ErrorIs(err, dto.Err_INTERNAL_FAILED_SCAN_USER)
}
