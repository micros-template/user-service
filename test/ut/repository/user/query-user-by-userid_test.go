package repository_test

import (
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"github.com/dropboks/sharedlib/model"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type GetUserByIdRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
}

func (g *GetUserByIdRepositorySuite) SetupSuite() {

	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	g.NoError(err)
	g.mockPgx = pgxMock
	g.userRepository = repository.NewUserRepository(pgxMock, logger)
}

func TestGetUserByIdRepositorySuite(t *testing.T) {
	suite.Run(t, &GetUserByIdRepositorySuite{})
}

func (g *GetUserByIdRepositorySuite) TestUserRepository_GetUserById_Success() {
	userId := "123"
	image := "image.png"
	expectedUser := &model.User{
		ID:               userId,
		FullName:         "John Doe",
		Image:            &image,
		Email:            "john@example.com",
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

	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE id = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(userId).WillReturnRows(rows)

	user, err := g.userRepository.QueryUserByUserId(userId)
	g.NoError(err)
	g.Equal(expectedUser, user)
}

func (g *GetUserByIdRepositorySuite) TestAuthRepository_GetUserById_NotFound() {
	userId := "notfound"
	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE id = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(userId).WillReturnError(pgx.ErrNoRows)

	user, err := g.userRepository.QueryUserByUserId(userId)
	g.Nil(user)
	g.ErrorIs(err, dto.Err_NOTFOUND_USER_NOT_FOUND)
}

func (g *GetUserByIdRepositorySuite) TestAuthRepository_GetUserById_ScanError() {
	userId := "123"
	rows := pgxmock.NewRows([]string{
		"id", "full_name", "image", "email", "password", "verified", "two_factor_enabled",
	}).AddRow(
		123, // should be string, but using int to cause scan error
		"John Doe",
		"image.png",
		"john@example.com",
		"hashedpassword",
		true,
		false,
	)
	query := `SELECT id, full_name, image, email, password, verified, two_factor_enabled FROM users WHERE id = \$1`
	g.mockPgx.ExpectQuery(query).WithArgs(userId).WillReturnRows(rows)

	user, err := g.userRepository.QueryUserByUserId(userId)
	g.Nil(user)
	g.ErrorIs(err, dto.Err_INTERNAL_FAILED_SCAN_USER)
}
