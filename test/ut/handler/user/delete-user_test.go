package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DeleteUserHandlerSuite struct {
	suite.Suite
	userHandler     handler.UserHandler
	mockUserService *mocks.UserServiceMock
	mockLogEmitter  *mocks.LoggerInfraMock
}

func (d *DeleteUserHandlerSuite) SetupSuite() {
	logger := zerolog.Nop()
	mockedUserService := new(mocks.UserServiceMock)
	mockedLogEmitter := new(mocks.LoggerInfraMock)
	d.mockUserService = mockedUserService
	d.mockLogEmitter = mockedLogEmitter
	d.userHandler = handler.NewUserHandler(mockedUserService, mockedLogEmitter, logger)
}

func (d *DeleteUserHandlerSuite) SetupTest() {
	d.mockUserService.ExpectedCalls = nil
	d.mockLogEmitter.ExpectedCalls = nil

	d.mockUserService.Calls = nil
	d.mockLogEmitter.Calls = nil
	gin.SetMode(gin.TestMode)
}

func TestDeleteUserHandlerSuite(t *testing.T) {
	suite.Run(t, &DeleteUserHandlerSuite{})
}

func (d *DeleteUserHandlerSuite) TestUserHandler_DeleteUser_Success() {
	userId := "12345"
	reqBody := `{"password":"secret"}`
	d.mockUserService.On("DeleteUser", mock.AnythingOfType("*dto.DeleteUserRequest"), userId).Return(nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodDelete, "/", strings.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	// Act
	d.userHandler.DeleteUser(ctx)

	// Assert
	d.Equal(http.StatusOK, w.Code)
	d.Contains(w.Body.String(), dto.SUCCESS_DELETE_USER)
}

func (d *DeleteUserHandlerSuite) TestUserHandler_DeleteUser_MissingUserId() {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodDelete, "/", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	d.mockLogEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	d.userHandler.DeleteUser(ctx)

	d.Equal(http.StatusUnauthorized, w.Code)
	d.Contains(w.Body.String(), dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())

	time.Sleep(time.Second)
	d.mockLogEmitter.AssertExpectations(d.T())
}

func (d *DeleteUserHandlerSuite) TestUserHandler_DeleteUser_InvalidInput() {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodDelete, "/", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	d.mockLogEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)
	d.userHandler.DeleteUser(ctx)

	d.Equal(http.StatusBadRequest, w.Code)
	d.Contains(w.Body.String(), "invalid input")

	time.Sleep(time.Second)
	d.mockLogEmitter.AssertExpectations(d.T())
}

func (d *DeleteUserHandlerSuite) TestUserHandler_DeleteUser_WrongPassword() {
	userId := "12345"
	reqBody := `{"password":"wrong-password"}`
	d.mockUserService.On("DeleteUser", mock.AnythingOfType("*dto.DeleteUserRequest"), userId).
		Return(dto.Err_UNAUTHORIZED_PASSWORD_WRONG)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodDelete, "/", strings.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	d.userHandler.DeleteUser(ctx)

	d.Equal(http.StatusUnauthorized, w.Code)
	d.Contains(w.Body.String(), dto.Err_UNAUTHORIZED_PASSWORD_WRONG.Error())
}

func (d *DeleteUserHandlerSuite) TestUserHandler_DeleteUser_UserNotFound() {
	userId := "12345"
	reqBody := `{"password":"secret"}`
	d.mockUserService.On("DeleteUser", mock.AnythingOfType("*dto.DeleteUserRequest"), userId).
		Return(dto.Err_NOTFOUND_USER_NOT_FOUND)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest(http.MethodDelete, "/", strings.NewReader(reqBody))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	d.userHandler.DeleteUser(ctx)

	d.Equal(http.StatusNotFound, w.Code)
	d.Contains(w.Body.String(), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
}
