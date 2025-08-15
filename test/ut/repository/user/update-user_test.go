package repository_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/micros-template/user-service/internal/domain/dto"
	"github.com/micros-template/user-service/internal/domain/repository"
	mk "github.com/micros-template/user-service/test/mocks"

	"github.com/micros-template/sharedlib/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
	logEmitter     *mk.LoggerInfraMock
}

func (u *UpdateUserRepositorySuite) SetupSuite() {
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	mockLogEmitter := new(mk.LoggerInfraMock)
	u.NoError(err)
	u.mockPgx = pgxMock
	u.logEmitter = mockLogEmitter
	u.userRepository = repository.NewUserRepository(pgxMock, mockLogEmitter, logger)
}

func (u *UpdateUserRepositorySuite) SetupTest() {
	u.logEmitter.ExpectedCalls = nil
	u.logEmitter.Calls = nil
}

func TestUpdateUserRepositorySuite(t *testing.T) {
	suite.Run(t, &UpdateUserRepositorySuite{})
}

func (u *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_Success() {
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
	u.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := u.userRepository.UpdateUser(user)
	u.NoError(err)
}

func (u *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_NotFound() {
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
	u.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	u.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userRepository.UpdateUser(user)
	u.ErrorIs(err, dto.Err_NOTFOUND_USER_NOT_FOUND)

	time.Sleep(time.Second)
	u.logEmitter.AssertExpectations(u.T())
}

func (u *UpdateUserRepositorySuite) TestUserRepository_UpdateUser_QueryError() {
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
	u.mockPgx.ExpectExec(updateQuery).
		WithArgs(user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled, user.ID).
		WillReturnError(fmt.Errorf("query execution failed"))
	u.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userRepository.UpdateUser(user)
	u.ErrorIs(err, dto.Err_INTERNAL_FAILED_UPDATE_USER)

	time.Sleep(time.Second)
	u.logEmitter.AssertExpectations(u.T())
}
