package handler_test

import (
	"bytes"
	"log"
	"mime/multipart"
	"testing"

	"net/http"
	"net/http/httptest"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UpdateUserHandlerSuite struct {
	suite.Suite
	userHandler     handler.UserHandler
	mockUserService *mocks.UserServiceMock
}

func (u *UpdateUserHandlerSuite) SetupSuite() {
	logger := zerolog.Nop()
	mockedUserService := new(mocks.UserServiceMock)
	u.mockUserService = mockedUserService
	u.userHandler = handler.NewUserHandler(mockedUserService, logger)
}

func (u *UpdateUserHandlerSuite) SetupTest() {
	u.mockUserService.ExpectedCalls = nil
	u.mockUserService.Calls = nil
	gin.SetMode(gin.TestMode)
}

func TestUpdateUserHandlerSuite(t *testing.T) {
	suite.Run(t, &UpdateUserHandlerSuite{})
}
func (u *UpdateUserHandlerSuite) TestUserHandler_UpdateUser_Success() {
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	formWriter.Close()

	request := httptest.NewRequest(http.MethodPatch, "/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = request

	u.mockUserService.On("UpdateUser", mock.Anything, "12345").Return(nil)
	u.userHandler.UpdateUser(ctx)

	u.Equal(http.StatusOK, w.Code)
	u.Contains(w.Body.String(), "success update profile data")

	u.mockUserService.AssertExpectations(u.T())
}

func (u *UpdateUserHandlerSuite) TestUserHandler_UpdateUser_MissingUserId() {

	request := httptest.NewRequest(http.MethodPatch, "/", nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = request

	u.userHandler.UpdateUser(ctx)

	u.Equal(http.StatusUnauthorized, w.Code)
	u.Contains(w.Body.String(), dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
}

func (u *UpdateUserHandlerSuite) TestUserHandler_UpdateUser_UserNotFound() {
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	formWriter.Close()

	request := httptest.NewRequest(http.MethodPatch, "/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = request

	u.mockUserService.On("UpdateUser", mock.Anything, "12345").Return(dto.Err_NOTFOUND_USER_NOT_FOUND)
	u.userHandler.UpdateUser(ctx)

	u.Equal(http.StatusNotFound, w.Code)
	u.Contains(w.Body.String(), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())

	u.mockUserService.AssertExpectations(u.T())
}

func (u *UpdateUserHandlerSuite) TestUserHandler_UpdateUser_ImageWrongExtension() {
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	fileWriter, _ := formWriter.CreateFormFile("image", "test.webp")
	_, err := fileWriter.Write([]byte("fake image data"))
	if err != nil {
		log.Fatal("failed to create image data")
	}
	formWriter.Close()

	request := httptest.NewRequest(http.MethodPatch, "/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = request

	u.mockUserService.On("UpdateUser", mock.Anything, "12345").Return(dto.Err_BAD_REQUEST_WRONG_EXTENSION)
	u.userHandler.UpdateUser(ctx)

	u.Equal(http.StatusBadRequest, w.Code)
	u.Contains(w.Body.String(), dto.Err_BAD_REQUEST_WRONG_EXTENSION.Error())

	u.mockUserService.AssertExpectations(u.T())
}

func (u *UpdateUserHandlerSuite) TestUserHandler_UpdateUser_ImageLimitSizeExceeded() {
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	fileWriter, _ := formWriter.CreateFormFile("image", "test.webp")
	_, err := fileWriter.Write([]byte("fake image data"))
	if err != nil {
		log.Fatal("failed to create image data")
	}
	formWriter.Close()

	request := httptest.NewRequest(http.MethodPatch, "/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = request

	u.mockUserService.On("UpdateUser", mock.Anything, "12345").Return(dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED)
	u.userHandler.UpdateUser(ctx)

	u.Equal(http.StatusBadRequest, w.Code)
	u.Contains(w.Body.String(), dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED.Error())

	u.mockUserService.AssertExpectations(u.T())
}
