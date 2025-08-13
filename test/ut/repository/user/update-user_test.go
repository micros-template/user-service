package repository_test

import (
	"fmt"
	"testing"
	"time"

	"10.1.20.130/dropping/log-management/pkg/mocks"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	mk "10.1.20.130/dropping/user-service/test/mocks"
	"github.com/dropboks/sharedlib/model"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
	mockUtil       *mk.LoggerServiceUtilMock
}

func (u *UpdateUserRepositorySuite) SetupSuite() {
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	u.mockUtil = new(mk.LoggerServiceUtilMock)
	mockLogEmitter := new(mocks.LogEmitterMock)
	u.NoError(err)
	u.mockPgx = pgxMock
	u.userRepository = repository.NewUserRepository(pgxMock, mockLogEmitter, u.mockUtil, logger)
}

func (u *UpdateUserRepositorySuite) SetupTest() {
	u.mockUtil.ExpectedCalls = nil
	u.mockUtil.Calls = nil
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
	u.mockUtil.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userRepository.UpdateUser(user)
	u.ErrorIs(err, dto.Err_NOTFOUND_USER_NOT_FOUND)

	time.Sleep(time.Second)
	u.mockUtil.AssertExpectations(u.T())
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
	u.mockUtil.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userRepository.UpdateUser(user)
	u.ErrorIs(err, dto.Err_INTERNAL_FAILED_UPDATE_USER)

	time.Sleep(time.Second)
	u.mockUtil.AssertExpectations(u.T())
}
