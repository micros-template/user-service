package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/handler"
	"10.1.20.130/dropping/user-service/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type ChangePasswordHandlerSuite struct {
	suite.Suite
	userHandler     handler.UserHandler
	mockUserService *mocks.UserServiceMock
}

func (c *ChangePasswordHandlerSuite) SetupSuite() {
	logger := zerolog.Nop()
	mockedUserService := new(mocks.UserServiceMock)
	c.mockUserService = mockedUserService
	c.userHandler = handler.NewUserHandler(mockedUserService, logger)
}

func (c *ChangePasswordHandlerSuite) SetupTest() {
	c.mockUserService.ExpectedCalls = nil
	c.mockUserService.Calls = nil
	gin.SetMode(gin.TestMode)
}

func TestChangePassowrdHandlerSuite(t *testing.T) {
	suite.Run(t, &ChangePasswordHandlerSuite{})
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_Success() {

	u := &dto.UpdatePasswordRequest{
		Password:           "old-password",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-password",
	}
	c.mockUserService.On("UpdatePassword", u, "12345").Return(nil)

	b := strings.NewReader(`{
		"password": "old-password",
		"new_password": "new-password",
		"confirm_new_password": "new-password"
	}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", b)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusOK, w.Code)
	c.Contains(w.Body.String(), "200")
	c.Contains(w.Body.String(), dto.SUCCESS_UPDATE_PASSWORD)
	c.mockUserService.AssertCalled(c.T(), "UpdatePassword", u, "12345")
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_MissingUserId() {

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusUnauthorized, w.Code)
	c.Contains(w.Body.String(), "401")
	c.Contains(w.Body.String(), "invalid token")
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_MissingBody() {

	b := strings.NewReader(`{
		"password": "old-password",
		"new_password": "new-password",
	}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", b)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Contains(w.Body.String(), "400")
	c.Contains(w.Body.String(), "invalid input")
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_PasswordAndConfirmPasswordNotMatch() {

	u := &dto.UpdatePasswordRequest{
		Password:           "old-password",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-passwor",
	}

	c.mockUserService.On("UpdatePassword", u, "12345").Return(dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH)

	b := strings.NewReader(`{
		"password": "old-password",
		"new_password": "new-password",
		"confirm_new_password": "new-passwor"
	}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", b)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Contains(w.Body.String(), "400")
	c.Contains(w.Body.String(), dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH.Error())
	c.mockUserService.AssertCalled(c.T(), "UpdatePassword", u, "12345")
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_WrongPassword() {

	u := &dto.UpdatePasswordRequest{
		Password:           "old-password",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-password",
	}

	c.mockUserService.On("UpdatePassword", u, "12345").Return(dto.Err_UNAUTHORIZED_PASSWORD_WRONG)

	b := strings.NewReader(`{
		"password": "old-password",
		"new_password": "new-password",
		"confirm_new_password": "new-password"
	}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", b)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusUnauthorized, w.Code)
	c.Contains(w.Body.String(), "401")
	c.Contains(w.Body.String(), dto.Err_UNAUTHORIZED_PASSWORD_WRONG.Error())
	c.mockUserService.AssertCalled(c.T(), "UpdatePassword", u, "12345")
}

func (c *ChangePasswordHandlerSuite) TestUserHandler_ChangePassword_UserNotFound() {

	u := &dto.UpdatePasswordRequest{
		Password:           "old-password",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-password",
	}

	c.mockUserService.On("UpdatePassword", u, "12345").Return(dto.Err_NOTFOUND_USER_NOT_FOUND)

	b := strings.NewReader(`{
		"password": "old-password",
		"new_password": "new-password",
		"confirm_new_password": "new-password"
	}`)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-password", b)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangePassword(ctx)

	c.Equal(http.StatusNotFound, w.Code)
	c.Contains(w.Body.String(), "404")
	c.Contains(w.Body.String(), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
	c.mockUserService.AssertCalled(c.T(), "UpdatePassword", u, "12345")
}
