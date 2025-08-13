package handler_test

import (
	"net/http"
	"net/http/httptest"
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

type GetProfileHandlerSuite struct {
	suite.Suite
	userHandler     handler.UserHandler
	mockUserService *mocks.UserServiceMock
	mockLogEmitter  *mocks.LoggerServiceUtilMock
}

func (g *GetProfileHandlerSuite) SetupSuite() {
	logger := zerolog.Nop()
	mockedUserService := new(mocks.UserServiceMock)
	mockedLogEmitter := new(mocks.LoggerServiceUtilMock)
	g.mockUserService = mockedUserService
	g.mockLogEmitter = mockedLogEmitter
	g.userHandler = handler.NewUserHandler(mockedUserService, mockedLogEmitter, logger)
}

func (g *GetProfileHandlerSuite) SetupTest() {
	g.mockUserService.ExpectedCalls = nil
	g.mockLogEmitter.ExpectedCalls = nil
	g.mockUserService.Calls = nil
	g.mockLogEmitter.Calls = nil
	gin.SetMode(gin.TestMode)
}

func TestGetProfileHandlerSuite(t *testing.T) {
	suite.Run(t, &GetProfileHandlerSuite{})
}
func (g *GetProfileHandlerSuite) TestUserHandler_GetProfile_Success() {
	res := dto.GetProfileResponse{
		FullName:         "test_user",
		Image:            new(string),
		Email:            "test@example.com",
		Verified:         false,
		TwoFactorEnabled: false,
	}
	g.mockUserService.On("GetProfile", "12345").Return(res, nil)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	g.userHandler.GetProfile(ctx)

	g.Equal(http.StatusOK, w.Code)
	g.Contains(w.Body.String(), "200")
	g.Contains(w.Body.String(), dto.SUCCESS_GET_PROFILE)
	g.mockUserService.AssertCalled(g.T(), "GetProfile", "12345")
}

func (g *GetProfileHandlerSuite) TestUserHandler_GetProfile_MissingUserId() {
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	g.mockLogEmitter.On("EmitLog", "ERR", mock.Anything).Return(nil)

	g.userHandler.GetProfile(ctx)

	g.Equal(http.StatusUnauthorized, w.Code)
	g.Contains(w.Body.String(), "401")
	g.Contains(w.Body.String(), dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())

	time.Sleep(time.Second)
	g.mockLogEmitter.AssertExpectations(g.T())
}

func (g *GetProfileHandlerSuite) TestUserHandler_GetProfile_UserNotFound() {
	g.mockUserService.On("GetProfile", "12345").Return(dto.GetProfileResponse{}, dto.Err_NOTFOUND_USER_NOT_FOUND)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/me", nil)
	ctx.Request.Header.Set("User-Data", `{"user_id":"12345"}`)

	g.userHandler.GetProfile(ctx)

	g.Equal(http.StatusNotFound, w.Code)
	g.Contains(w.Body.String(), "404")
	g.Contains(w.Body.String(), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
	g.mockUserService.AssertCalled(g.T(), "GetProfile", "12345")
}
