package repository_test

import (
	"regexp"
	"testing"
	"time"

	"10.1.20.130/dropping/log-management/pkg/mocks"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	mk "10.1.20.130/dropping/user-service/test/mocks"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DeleteUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
	mockUtil       *mk.UserServiceUtilMock
}

func (d *DeleteUserRepositorySuite) SetupSuite() {
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	mockUtil := new(mk.UserServiceUtilMock)
	mockLogEmitter := new(mocks.LogEmitterMock)

	d.NoError(err)
	d.mockUtil = mockUtil
	d.mockPgx = pgxMock
	d.userRepository = repository.NewUserRepository(pgxMock, mockLogEmitter, mockUtil, logger)
}

func (d *DeleteUserRepositorySuite) SetupTest() {
	d.mockUtil.ExpectedCalls = nil
	d.mockUtil.Calls = nil
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
	d.mockUtil.On("EmitLog", mock.Anything, "ERR", mock.Anything).Return(nil)

	err := d.userRepository.DeleteUser(userId)
	d.Equal(dto.Err_NOTFOUND_USER_NOT_FOUND, err)

	time.Sleep(time.Second)
	d.mockUtil.AssertExpectations(d.T())
}
