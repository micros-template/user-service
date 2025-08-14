package repository_test

import (
	"fmt"
	"testing"
	"time"

	"10.1.20.130/dropping/log-management/pkg/mocks"
	"10.1.20.130/dropping/sharedlib/model"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	mk "10.1.20.130/dropping/user-service/test/mocks"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CreateNewUserRepositorySuite struct {
	suite.Suite
	userRepository repository.UserRepository
	mockPgx        pgxmock.PgxPoolIface
	mockUtil       *mk.LoggerServiceUtilMock
}

func (u *CreateNewUserRepositorySuite) SetupSuite() {
	logger := zerolog.Nop()
	pgxMock, err := pgxmock.NewPool()
	mockUtil := new(mk.LoggerServiceUtilMock)
	mockLogEmitter := new(mocks.LogEmitterMock)
	u.NoError(err)
	u.mockPgx = pgxMock
	u.mockUtil = mockUtil
	u.userRepository = repository.NewUserRepository(pgxMock, mockLogEmitter, mockUtil, logger)
}

func (u *CreateNewUserRepositorySuite) SetupTest() {
	u.mockUtil.ExpectedCalls = nil
	u.mockUtil.Calls = nil
}

func TestCreateNewUserRepositorySuite(t *testing.T) {
	suite.Run(t, &CreateNewUserRepositorySuite{})
}

func (u *CreateNewUserRepositorySuite) TestUserRepository_CreateNewUser_Success() {
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
	u.mockPgx.ExpectQuery(insertQuery).
		WithArgs(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(user.ID))

	err := u.userRepository.CreateNewUser(user)
	u.NoError(err)
	u.Equal(email, user.ID)
}

func (u *CreateNewUserRepositorySuite) TestUserRepository_CreateNewUser_ScanError() {
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
	u.mockPgx.ExpectQuery(insertQuery).
		WithArgs(user.ID, user.FullName, user.Image, user.Email, user.Password, user.Verified, user.TwoFactorEnabled).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(struct{}{}))
	u.mockUtil.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userRepository.CreateNewUser(user)
	u.Error(err)
	u.ErrorIs(err, dto.Err_INTERNAL_FAILED_INSERT_USER)
	time.Sleep(time.Second)

	u.mockUtil.AssertExpectations(u.T())
}
