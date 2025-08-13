package it

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/test/helper"
	_helper "github.com/dropboks/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPGetProfileITSuite struct {
	suite.Suite
	ctx context.Context

	network                      *testcontainers.DockerNetwork
	gatewayContainer             *_helper.GatewayContainer
	userPgContainer              *_helper.PostgresContainer
	authPgContainer              *_helper.PostgresContainer
	redisContainer               *_helper.RedisContainer
	minioContainer               *_helper.MinioContainer
	natsContainer                *_helper.NatsContainer
	authContainer                *_helper.AuthServiceContainer
	userServiceContainer         *_helper.UserServiceContainer
	fileServiceContainer         *_helper.FileServiceContainer
	notificationServiceContainer *_helper.NotificationServiceContainer
	mailHogContainer             *_helper.MailhogContainer
}

func (g *HTTPGetProfileITSuite) SetupSuite() {

	log.Println("Setting up integration test suite for HTTPGetProfileITSuite")
	g.ctx = context.Background()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config")
	}
	// spawn sharedNetwork
	g.network = _helper.StartNetwork(g.ctx)

	// spawn user db
	userPgContainer, err := _helper.StartPostgresContainer(g.ctx, g.network.Name, "test_user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	g.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(g.ctx, g.network.Name, "test_auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	g.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartRedisContainer(g.ctx, g.network.Name, viper.GetString("container.redis_version"))
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	g.redisContainer = rContainer

	mContainer, err := _helper.StartMinioContainer(g.ctx, g.network.Name, viper.GetString("container.minio_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	g.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartNatsContainer(g.ctx, g.network.Name, viper.GetString("container.nats_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	g.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(g.ctx, g.network.Name, viper.GetString("container.auth_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting auth service container: %s", err)
	}
	g.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(g.ctx, g.network.Name, viper.GetString("container.file_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting file service container: %s", err)
	}
	g.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(g.ctx, g.network.Name, viper.GetString("container.user_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting user service container: %s", err)
	}
	g.userServiceContainer = uContainer

	noContainer, err := _helper.StartNotificationServiceContainer(g.ctx, g.network.Name, viper.GetString("container.notification_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting notification service container: %s", err)
	}
	g.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailhogContainer(g.ctx, g.network.Name, viper.GetString("container.mailhog_version"))
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	g.mailHogContainer = mailContainer

	gatewayContainer, err := _helper.StartGatewayContainer(g.ctx, g.network.Name, viper.GetString("container.gateway_version"))
	if err != nil {
		log.Fatalf("failed starting gateway container: %s", err)
	}
	g.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (g *HTTPGetProfileITSuite) TearDownSuite() {
	if err := g.userPgContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating user postgres container: %s", err)
	}
	if err := g.authPgContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating auth postgres container: %s", err)
	}
	if err := g.redisContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
	if err := g.minioContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating minio container: %s", err)
	}
	if err := g.natsContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating nats container: %s", err)
	}
	if err := g.authContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating auth service container: %s", err)
	}
	if err := g.userServiceContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := g.fileServiceContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := g.notificationServiceContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating notification service container: %s", err)
	}
	if err := g.mailHogContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating mailhog container: %s", err)
	}
	if err := g.gatewayContainer.Terminate(g.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for HTTPGetProfileITSuite")

}
func TestHTTPGetProfileITSuite(t *testing.T) {
	suite.Run(t, &HTTPGetProfileITSuite{})
}

func (g *HTTPGetProfileITSuite) TestGetProfileIT_Success() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, g.T())

	client := http.Client{}
	response, err := client.Do(request)
	g.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	g.NoError(err)

	g.Equal(http.StatusCreated, response.StatusCode)
	g.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		g.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", g.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	g.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	g.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	g.NoError(err)

	g.Equal(http.StatusOK, verifyResponse.StatusCode)
	g.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, g.T())

	client = http.Client{}
	response, err = client.Do(request)
	g.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	g.Equal(http.StatusOK, response.StatusCode)
	g.NoError(err)
	g.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	g.NoError(err)

	jwt, ok := respData["data"].(string)
	g.True(ok, "expected jwt token in data field")

	// get profile
	request, err = http.NewRequest(http.MethodGet, "http://localhost:9090/api/v1/user/me", nil)
	request.Header.Set("Authorization", "Bearer "+jwt)

	g.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	g.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	g.Equal(http.StatusOK, response.StatusCode)
	g.NoError(err)
	g.Contains(string(byteBody), "success get profile data")
}

func (g *HTTPGetProfileITSuite) TestGetProfileIT_MissingToken() {
	// get profile
	request, err := http.NewRequest(http.MethodGet, "http://localhost:9090/api/v1/user/me", nil)
	g.NoError(err)

	client := http.Client{}
	response, err := client.Do(request)
	g.NoError(err)

	g.Equal(http.StatusUnauthorized, response.StatusCode)
	g.NoError(err)
}

// not applicable due to checking in the previous layer and case (register, login, and verification middleware)

// func (g *HTTPGetProfileITSuite) TestGetProfileIT_UserNotFound() {
// 	// get profile
// 	request, err := http.NewRequest(http.MethodGet, "http://localhost:9090/api/v1/user/me", nil)
// 	g.NoError(err)
// 	request.Header.Set("User-Data", `{"user_id":"12345"}`)

// 	g.NoError(err)

// 	client := http.Client{}
// 	response, err := client.Do(request)
// 	g.NoError(err)

// 	byteBody, err := io.ReadAll(response.Body)

// 	g.Equal(http.StatusNotFound, response.StatusCode)
// 	g.NoError(err)
// 	g.Contains(string(byteBody), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
// }
