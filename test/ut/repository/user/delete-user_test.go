package repository_test

import (
	"regexp"
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type DeleteUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
}

func (d *DeleteUserRepositorySuite) SetupSuite() {
	// logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	d.NoError(err)
	d.mockPgx = pgxMock
	d.userRepository = repository.NewUserRepository(pgxMock, logger)
}

func TestDeleteUserRepositorySuite(t *testing.T) {
	suite.Run(t, &DeleteUserRepositorySuite{})
}

func (d *DeleteUserRepositorySuite) TestUserRepository_DeleteUser_Success() {
	userId := "user-123"
	query := `DELETE FROM users WHERE id = $1`
	d.mockPgx.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(userId).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := d.userRepository.DeleteUser(userId)
	d.NoError(err)
}
func (d *DeleteUserRepositorySuite) TestUserRepository_DeleteUser_UserNotFound() {
	userId := "user-456"
	query := `DELETE FROM users WHERE id = $1`
	d.mockPgx.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(userId).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := d.userRepository.DeleteUser(userId)
	d.Equal(dto.Err_NOTFOUND_USER_NOT_FOUND, err)
}
