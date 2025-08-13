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

type ChangeEmailHandlerSuite struct {
	suite.Suite
	userHandler     handler.UserHandler
	mockUserService *mocks.UserServiceMock
	mockLogEmitter  *mocks.LoggerServiceUtilMock
}

func (c *ChangeEmailHandlerSuite) SetupSuite() {
	logger := zerolog.Nop()
	mockedUserService := new(mocks.UserServiceMock)
	mockedLogEmitterService := new(mocks.LoggerServiceUtilMock)
	c.mockUserService = mockedUserService
	c.mockLogEmitter = mockedLogEmitterService
	c.userHandler = handler.NewUserHandler(mockedUserService, mockedLogEmitterService, logger)
}

func (c *ChangeEmailHandlerSuite) SetupTest() {
	c.mockUserService.ExpectedCalls = nil
	c.mockLogEmitter.ExpectedCalls = nil
	c.mockUserService.Calls = nil
	c.mockLogEmitter.Calls = nil
	gin.SetMode(gin.TestMode)
}

func TestChangeEmailHandlerSuite(t *testing.T) {
	suite.Run(t, &ChangeEmailHandlerSuite{})
}

func (c *ChangeEmailHandlerSuite) TestUserHandler_ChangeEmail_Success() {

	c.mockUserService.On("UpdateEmail", &dto.UpdateEmailRequest{Email: "newemail@example.com"}, "12345").Return(nil)

	body := `{"email":"newemail@example.com"}`

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-email",
		strings.NewReader(body))
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")

	c.userHandler.ChangeEmail(ctx)

	c.Equal(http.StatusOK, w.Code)
	c.Contains(w.Body.String(), "200")
	c.Contains(w.Body.String(), dto.SUCCESS_UPDATE_EMAIL)

	c.mockUserService.AssertCalled(c.T(), "UpdateEmail", &dto.UpdateEmailRequest{Email: "newemail@example.com"}, "12345")
}

func (c *ChangeEmailHandlerSuite) TestUserHandler_ChangeEmail_MissingUserId() {

	body := `{"email":"newemail@example.com"}`

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-email", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	c.mockLogEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)
	c.userHandler.ChangeEmail(ctx)

	c.Equal(http.StatusUnauthorized, w.Code)
	c.Contains(w.Body.String(), "401")
	c.Contains(w.Body.String(), "invalid token")

	time.Sleep(time.Second)
	c.mockLogEmitter.AssertExpectations(c.T())
}

func (c *ChangeEmailHandlerSuite) TestUserHandler_ChangeEmail_MissingBody() {

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/change-email", nil)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)
	ctx.Request.Header.Set("Content-Type", "application/json")
	c.mockLogEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	c.userHandler.ChangeEmail(ctx)

	c.Equal(http.StatusBadRequest, w.Code)
	c.Contains(w.Body.String(), "400")
	c.Contains(w.Body.String(), "invalid input")

	time.Sleep(time.Second)
	c.mockLogEmitter.AssertExpectations(c.T())
}
