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

type CreateNewUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
}

func (g *CreateNewUserRepositorySuite) SetupSuite() {
	// logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	g.NoError(err)
	g.mockPgx = pgxMock
	g.userRepository = repository.NewUserRepository(pgxMock, logger)
}

func TestCreateNewUserRepositorySuite(t *testing.T) {
	suite.Run(t, &CreateNewUserRepositorySuite{})
}

func (g *CreateNewUserRepositorySuite) TestUserRepository_CreateNewUser_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	image := "image.png"
	user := &model.User{
		ID:               email,
		FullName:         "test_user",
		Image:            &image,
		Email:            email,
		Password:         "hashedpassword",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	insertQuery := `INSERT INTO users \(id,full_name,image,email,password,verified,two_factor_enabled\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING id`
	g.mockPgx.ExpectQuery(insertQuery).
		WithArgs(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(user.ID))

	err := g.userRepository.CreateNewUser(user)
	g.NoError(err)
	g.Equal(email, user.ID)
}

func (g *CreateNewUserRepositorySuite) TestUserRepository_CreateNewUser_NotFound() {
	email := "notfound@example.com"
	image := "image.png"
	user := &model.User{
		ID:               email,
		FullName:         "test_user",
		Image:            &image,
		Email:            email,
		Password:         "hashedpassword",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	insertQuery := `INSERT INTO users \(id,full_name,image,email,password,verified,two_factor_enabled\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING id`
	g.mockPgx.ExpectQuery(insertQuery).
		WithArgs(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		WillReturnError(pgx.ErrNoRows)

	err := g.userRepository.CreateNewUser(user)
	g.ErrorIs(err, dto.Err_INTERNAL_FAILED_INSERT_USER)
}

func (g *CreateNewUserRepositorySuite) TestUserRepository_CreateNewUser_ScanError() {
	email := "scanerror@example.com"
	image := "image.png"
	user := &model.User{
		ID:               email,
		FullName:         "test_user",
		Image:            &image,
		Email:            email,
		Password:         "hashedpassword",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	insertQuery := `INSERT INTO users \(id,full_name,image,email,password,verified,two_factor_enabled\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING id`
	g.mockPgx.ExpectQuery(insertQuery).
		WithArgs(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(struct{}{}))

	err := g.userRepository.CreateNewUser(user)
	g.Error(err)
	fmt.Println("ScanError test got error:", err)
	g.ErrorIs(err, dto.Err_INTERNAL_FAILED_INSERT_USER)
}
