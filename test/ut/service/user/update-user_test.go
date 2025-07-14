package service_test

import (
	"bytes"
	"mime/multipart"
	"testing"
	"time"

	"github.com/dropboks/proto-file/pkg/fpb"
	"github.com/dropboks/sharedlib/model"
	"github.com/dropboks/user-service/internal/domain/dto"
	"github.com/dropboks/user-service/internal/domain/service"
	"github.com/dropboks/user-service/test/mocks"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateUserUserServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mocks.UserRepositoryMock
	eventEmitter       *mocks.EmitterMock
	fileService        *mocks.MockFileServiceClient
	notificationStream *mocks.MockNatsInfra
	redisRepository    *mocks.MockRedisRepository
}

func (u *UpdateUserUserServiceSuite) SetupSuite() {

	mockUserRepo := new(mocks.UserRepositoryMock)
	mockEventEmitter := new(mocks.EmitterMock)
	mockFileService := new(mocks.MockFileServiceClient)
	mockNotificationStream := new(mocks.MockNatsInfra)
	mockRedisRepository := new(mocks.MockRedisRepository)

	logger := zerolog.Nop()
	// logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	u.userRepository = mockUserRepo
	u.eventEmitter = mockEventEmitter
	u.fileService = mockFileService
	u.notificationStream = mockNotificationStream
	u.redisRepository = mockRedisRepository
	u.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter)
}

func (u *UpdateUserUserServiceSuite) SetupTest() {
	u.userRepository.ExpectedCalls = nil
	u.eventEmitter.ExpectedCalls = nil
	u.fileService.ExpectedCalls = nil
	u.notificationStream.ExpectedCalls = nil
	u.redisRepository.ExpectedCalls = nil

	u.userRepository.Calls = nil
	u.eventEmitter.Calls = nil
	u.fileService.Calls = nil
	u.notificationStream.Calls = nil
	u.redisRepository.Calls = nil

}

func TestUpdateUserUserServiceSuite(t *testing.T) {
	suite.Run(t, &UpdateUserUserServiceSuite{})
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_Success() {
	userId := "user-123"
	req := &dto.UpdateUserRequest{
		FullName:         "Updated Name",
		TwoFactorEnabled: true,
	}
	user := &model.User{
		ID:               userId,
		FullName:         "Original Name",
		TwoFactorEnabled: false,
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(user, nil)
	u.userRepository.On("UpdateUser", mock.Anything).Return(nil)
	u.eventEmitter.On("UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User")).Return(nil).Maybe()

	err := u.userService.UpdateUser(req, userId)

	u.NoError(err)
	u.userRepository.AssertExpectations(u.T())

	time.Sleep(time.Second)
	u.eventEmitter.AssertCalled(u.T(), "UpdateUser", mock.Anything, mock.AnythingOfType("*upb.User"))
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_UserNotFound() {
	userId := "user-404"
	req := &dto.UpdateUserRequest{
		FullName: "Nonexistent User",
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(nil, dto.Err_NOTFOUND_USER_NOT_FOUND)

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.Equal(dto.Err_NOTFOUND_USER_NOT_FOUND, err)
	u.userRepository.AssertExpectations(u.T())
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_InvalidImageExtension() {
	userId := "user-123"
	req := &dto.UpdateUserRequest{
		Image: &multipart.FileHeader{
			Filename: "invalid.bmp",
		},
	}
	user := &model.User{
		ID: userId,
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(user, nil)

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.Equal(dto.Err_BAD_REQUEST_WRONG_EXTENSION, err)
	u.userRepository.AssertExpectations(u.T())
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_ImageUploadError() {
	userId := "user-123"

	imageData := bytes.Repeat([]byte("test"), 1024)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("image", "valid.jpg")
	part.Write(imageData)
	writer.Close()

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)
	fileHeader := form.File["image"][0]

	req := &dto.UpdateUserRequest{
		Image: fileHeader,
	}
	user := &model.User{
		ID: userId,
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(user, nil)

	imageReq := &fpb.Image{
		Image: imageData,
		Ext:   "jpg",
	}
	u.fileService.On("SaveProfileImage", mock.Anything, imageReq).Return(nil, status.Errorf(codes.Internal, "upload error"))

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.userRepository.AssertExpectations(u.T())
	u.fileService.AssertExpectations(u.T())
}
