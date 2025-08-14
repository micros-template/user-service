package it

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	_helper "10.1.20.130/dropping/sharedlib/test/helper"
	"10.1.20.130/dropping/user-service/test/helper"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPChangeEmailITSuite struct {
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

func (c *HTTPChangeEmailITSuite) SetupSuite() {
	log.Println("Setting up integration test suite for HTTPChangeEmailITSuite")
	c.ctx = context.Background()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config")
	}
	// spawn sharedNetwork
	c.network = _helper.StartNetwork(c.ctx)

	// spawn user db
	userPgContainer, err := _helper.StartPostgresContainer(c.ctx, c.network.Name, "test_user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	c.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(c.ctx, c.network.Name, "test_auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	c.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartRedisContainer(c.ctx, c.network.Name, viper.GetString("container.redis_version"))
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	c.redisContainer = rContainer

	mContainer, err := _helper.StartMinioContainer(c.ctx, c.network.Name, viper.GetString("container.minio_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	c.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartNatsContainer(c.ctx, c.network.Name, viper.GetString("container.nats_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	c.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(c.ctx, c.network.Name, viper.GetString("container.auth_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting auth service container: %s", err)
	}
	c.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(c.ctx, c.network.Name, viper.GetString("container.file_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting file service container: %s", err)
	}
	c.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(c.ctx, c.network.Name, viper.GetString("container.user_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting user service container: %s", err)
	}
	c.userServiceContainer = uContainer

	noContainer, err := _helper.StartNotificationServiceContainer(c.ctx, c.network.Name, viper.GetString("container.notification_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting notification service container: %s", err)
	}
	c.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailhogContainer(c.ctx, c.network.Name, viper.GetString("container.mailhog_version"))
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	c.mailHogContainer = mailContainer

	gatewayContainer, err := _helper.StartGatewayContainer(c.ctx, c.network.Name, viper.GetString("container.gateway_version"))
	if err != nil {
		log.Fatalf("failed starting gateway container: %s", err)
	}
	c.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (c *HTTPChangeEmailITSuite) TearDownSuite() {
	if err := c.userPgContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating user postgres container: %s", err)
	}
	if err := c.authPgContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating auth postgres container: %s", err)
	}
	if err := c.redisContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
	if err := c.minioContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating minio container: %s", err)
	}
	if err := c.natsContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating nats container: %s", err)
	}
	if err := c.authContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating auth service container: %s", err)
	}
	if err := c.userServiceContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := c.fileServiceContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := c.notificationServiceContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating notification service container: %s", err)
	}
	if err := c.mailHogContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating mailhog container: %s", err)
	}
	if err := c.gatewayContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for HTTPChangeEmailITSuite")

}
func TestHTTPChangeEmailITSuite(t *testing.T) {
	suite.Run(t, &HTTPChangeEmailITSuite{})
}

func (c *HTTPChangeEmailITSuite) TestChangeEmailIT_Success() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, c.T())

	client := http.Client{}
	response, err := client.Do(request)
	c.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	c.NoError(err)

	c.Equal(http.StatusCreated, response.StatusCode)
	c.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		c.T().Errorf("error closing response body: %v", err)
	}

	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", c.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	c.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	c.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	c.NoError(err)

	c.Equal(http.StatusOK, verifyResponse.StatusCode)
	c.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, c.T())

	client = http.Client{}
	response, err = client.Do(request)
	c.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(err)
	c.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	c.NoError(err)

	jwt, ok := respData["data"].(string)
	c.True(ok, "expected jwt token in data field")

	// change email
	reqBody := &bytes.Buffer{}

	encoder := gin.H{
		"email": email,
	}
	_ = json.NewEncoder(reqBody).Encode(encoder)

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/email", reqBody)
	request.Header.Set("Authorization", "Bearer "+jwt)
	c.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	c.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(err)
	c.Contains(string(byteBody), "verify to change email")
}

// not applicable because verify is in middlewware
func (c *HTTPChangeEmailITSuite) TestChangeEmailIT_MisingToken() {

	request, err := http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/email", nil)
	c.NoError(err)

	client := http.Client{}
	response, err := client.Do(request)
	c.NoError(err)

	c.Equal(http.StatusUnauthorized, response.StatusCode)
	c.NoError(err)
}

func (c *HTTPChangeEmailITSuite) TestChangeEmailIT_MissingBody() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, c.T())

	client := http.Client{}
	response, err := client.Do(request)
	c.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	c.NoError(err)

	c.Equal(http.StatusCreated, response.StatusCode)
	c.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		c.T().Errorf("error closing response body: %v", err)
	}

	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", c.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	c.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	c.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	c.NoError(err)

	c.Equal(http.StatusOK, verifyResponse.StatusCode)
	c.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, c.T())

	client = http.Client{}
	response, err = client.Do(request)
	c.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(err)
	c.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	c.NoError(err)

	jwt, ok := respData["data"].(string)
	c.True(ok, "expected jwt token in data field")

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/email", nil)
	request.Header.Set("Authorization", "Bearer "+jwt)

	c.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	c.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	c.Equal(http.StatusBadRequest, response.StatusCode)
	c.NoError(err)
	c.Contains(string(byteBody), "invalid input")
}
