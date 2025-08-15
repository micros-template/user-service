package service_test

import (
	"bytes"
	"log"
	"mime/multipart"
	"testing"
	"time"

	"github.com/micros-template/user-service/internal/domain/dto"
	"github.com/micros-template/user-service/internal/domain/service"
	mk "github.com/micros-template/user-service/test/mocks"

	"github.com/micros-template/proto-file/pkg/fpb"
	"github.com/micros-template/sharedlib/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpdateUserUserServiceSuite struct {
	suite.Suite
	userService        service.UserService
	userRepository     *mk.UserRepositoryMock
	eventEmitter       *mk.EmitterMock
	fileService        *mk.MockFileServiceClient
	notificationStream *mk.MockNatsInfra
	redisRepository    *mk.MockRedisRepository
	logEmitter         *mk.LoggerInfraMock
}

func (u *UpdateUserUserServiceSuite) SetupSuite() {

	mockUserRepo := new(mk.UserRepositoryMock)
	mockEventEmitter := new(mk.EmitterMock)
	mockFileService := new(mk.MockFileServiceClient)
	mockNotificationStream := new(mk.MockNatsInfra)
	mockRedisRepository := new(mk.MockRedisRepository)
	mockLogEmitter := new(mk.LoggerInfraMock)

	logger := zerolog.Nop()
	u.userRepository = mockUserRepo
	u.eventEmitter = mockEventEmitter
	u.fileService = mockFileService
	u.notificationStream = mockNotificationStream
	u.redisRepository = mockRedisRepository
	u.logEmitter = mockLogEmitter
	u.userService = service.NewUserService(mockUserRepo, logger, mockFileService, mockRedisRepository, mockNotificationStream, mockEventEmitter, mockLogEmitter)
}

func (u *UpdateUserUserServiceSuite) SetupTest() {
	u.userRepository.ExpectedCalls = nil
	u.eventEmitter.ExpectedCalls = nil
	u.fileService.ExpectedCalls = nil
	u.notificationStream.ExpectedCalls = nil
	u.redisRepository.ExpectedCalls = nil
	u.logEmitter.ExpectedCalls = nil

	u.userRepository.Calls = nil
	u.eventEmitter.Calls = nil
	u.fileService.Calls = nil
	u.notificationStream.Calls = nil
	u.redisRepository.Calls = nil
	u.logEmitter.Calls = nil
}

func TestUpdateUserUserServiceSuite(t *testing.T) {
	suite.Run(t, &UpdateUserUserServiceSuite{})
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_Success() {
	userId := "user-123"
	req := &dto.UpdateUserRequest{
		FullName:         "Updated Name",
		TwoFactorEnabled: true,
		Image:            nil,
	}
	user := &model.User{
		ID:               userId,
		FullName:         "Original Name",
		TwoFactorEnabled: false,
		Image:            nil,
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
	u.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.Equal(dto.Err_BAD_REQUEST_WRONG_EXTENSION, err)
	u.userRepository.AssertExpectations(u.T())

	time.Sleep(time.Second)
	u.logEmitter.AssertExpectations(u.T())
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_SizeLimitExceeded() {
	imageData := bytes.Repeat([]byte("test"), 8*1024*1024)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("image", "test_image.jpg")
	if _, err := part.Write(imageData); err != nil {
		log.Fatal("failed to write image data:", err)
	}
	if err := writer.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}

	reader := multipart.NewReader(&buf, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)
	fileHeader := form.File["image"][0]
	userId := "user-123"
	req := &dto.UpdateUserRequest{
		Image: fileHeader,
	}
	user := &model.User{
		ID: userId,
	}
	u.userRepository.On("QueryUserByUserId", userId).Return(user, nil)
	u.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.Equal(dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED, err)
	u.userRepository.AssertExpectations(u.T())

	time.Sleep(time.Second)
	u.logEmitter.AssertExpectations(u.T())
}

func (u *UpdateUserUserServiceSuite) TestUserService_UpdateUser_ImageUploadError() {
	userId := "user-123"

	imageData := bytes.Repeat([]byte("test"), 1024)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("image", "valid.jpg")
	if _, err := part.Write(imageData); err != nil {
		log.Fatal("failed to write image data:", err)
	}
	if err := writer.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}

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
	u.logEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	err := u.userService.UpdateUser(req, userId)

	u.Error(err)
	u.userRepository.AssertExpectations(u.T())
	u.fileService.AssertExpectations(u.T())

	time.Sleep(time.Second)
	u.logEmitter.AssertExpectations(u.T())
}
