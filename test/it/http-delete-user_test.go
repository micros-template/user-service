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

	"github.com/micros-template/user-service/internal/domain/dto"
	"github.com/micros-template/user-service/test/helper"

	"github.com/gin-gonic/gin"
	_helper "github.com/micros-template/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPDeleteUserITSuite struct {
	suite.Suite
	ctx context.Context

	network                      *testcontainers.DockerNetwork
	gatewayContainer             *_helper.GatewayContainer
	userPgContainer              *_helper.SQLContainer
	authPgContainer              *_helper.SQLContainer
	redisContainer               *_helper.CacheContainer
	minioContainer               *_helper.StorageContainer
	natsContainer                *_helper.MessageQueueContainer
	authContainer                *_helper.AuthServiceContainer
	userServiceContainer         *_helper.UserServiceContainer
	fileServiceContainer         *_helper.FileServiceContainer
	notificationServiceContainer *_helper.NotificationServiceContainer
	mailHogContainer             *_helper.MailContainer
}

func (d *HTTPDeleteUserITSuite) SetupSuite() {

	log.Println("Setting up integration test suite for HTTPDeleteUserITSuite")
	d.ctx = context.Background()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config")
	}
	// spawn sharedNetwork
	d.network = _helper.StartNetwork(d.ctx)

	// spawn user db
	userPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 d.ctx,
		SharedNetwork:           d.network.Name,
		ImageName:               viper.GetString("container.postgresql_image"),
		ContainerName:           "test_user_db",
		SQLInitScriptPath:       viper.GetString("script.init_sql"),
		SQLInitInsideScriptPath: "/docker-entrypoint-initdb.d/init-db.sql",
		WaitingSignal:           "database system is ready to accept connections",
		Env: map[string]string{
			"POSTGRES_DB":       viper.GetString("database.name"),
			"POSTGRES_USER":     viper.GetString("database.user"),
			"POSTGRES_PASSWORD": viper.GetString("database.password"),
		},
	})
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	d.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 d.ctx,
		SharedNetwork:           d.network.Name,
		ImageName:               viper.GetString("container.postgresql_image"),
		ContainerName:           "test_auth_db",
		SQLInitScriptPath:       viper.GetString("script.init_sql"),
		SQLInitInsideScriptPath: "/docker-entrypoint-initdb.d/init-db.sql",
		WaitingSignal:           "database system is ready to accept connections",
		Env: map[string]string{
			"POSTGRES_DB":       viper.GetString("database.name"),
			"POSTGRES_USER":     viper.GetString("database.user"),
			"POSTGRES_PASSWORD": viper.GetString("database.password"),
		},
	})
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	d.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartCacheContainer(_helper.CacheParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.redis_image"),
		ContainerName: "test_redis",
		WaitingSignal: "6379/tcp",
		Cmd:           []string{"redis-server", "--requirepass", viper.GetString("redis.password")},
		Env: map[string]string{
			"REDIS_PASSWORD": viper.GetString("redis.password"),
		},
	})
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	d.redisContainer = rContainer

	mContainer, err := _helper.StartStorageContainer(_helper.StorageParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.minio_image"),
		ContainerName: "test-minio",
		WaitingSignal: "API:",
		Cmd:           []string{"server", "/data"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     viper.GetString("minio.credential.user"),
			"MINIO_ROOT_PASSWORD": viper.GetString("minio.credential.password"),
		},
	})
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	d.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartMessageQueueContainer(_helper.MessageQueueParameterOption{
		Context:            d.ctx,
		SharedNetwork:      d.network.Name,
		ImageName:          viper.GetString("container.nats_image"),
		ContainerName:      "test_nats",
		MQConfigPath:       viper.GetString("script.nats_server"),
		MQInsideConfigPath: "/etc/nats/nats.conf",
		WaitingSignal:      "Server is ready",
		MappedPort:         []string{"4221:4221/tcp"},
		Cmd: []string{
			"-c", "/etc/nats/nats.conf",
			"--name", "nats",
			"-p", "4221",
		},
		Env: map[string]string{
			"NATS_USER":     viper.GetString("nats.credential.user"),
			"NATS_PASSWORD": viper.GetString("nats.credential.password"),
		},
	})
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	d.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(_helper.AuthServiceParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.auth_service_image"),
		ContainerName: "test_auth_service",
		WaitingSignal: "HTTP Server Starting in port",
		Cmd:           []string{"/auth_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting auth service container: %s", err)
	}
	d.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(_helper.FileServiceParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.file_service_image"),
		ContainerName: "test_file_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/file_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting file service container: %s", err)
	}
	d.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(_helper.UserServiceParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.user_service_image"),
		ContainerName: "test_user_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/user_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting user service container: %s", err)
	}
	d.userServiceContainer = uContainer

	noContainer, err := _helper.StartNotificationServiceContainer(_helper.NotificationServiceParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.notification_service_image"),
		ContainerName: "test_notification_service",
		WaitingSignal: "subscriber for notification is running",
		Cmd:           []string{"/notification_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting notification service container: %s", err)
	}
	d.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailContainer(_helper.MailParameterOption{
		Context:       d.ctx,
		SharedNetwork: d.network.Name,
		ImageName:     viper.GetString("container.mailhog_image"),
		ContainerName: "mailhog",
		WaitingSignal: "1025/tcp",
		MappedPort:    []string{"1025:1025/tcp", "8025:8025/tcp"},
	})
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	d.mailHogContainer = mailContainer

	gatewayContainer, err := _helper.StartGatewayContainer(_helper.GatewayParameterOption{
		Context:                   d.ctx,
		SharedNetwork:             d.network.Name,
		ImageName:                 viper.GetString("container.gateway_image"),
		ContainerName:             "test_gateway",
		NginxConfigPath:           viper.GetString("script.nginx"),
		NginxInsideConfigPath:     "/etc/nginx/conf.d/default.conf",
		GrpcErrorConfigPath:       viper.GetString("script.grpc_error"),
		GrpcErrorInsideConfigPath: "/etc/nginx/conf.d/errors.grpc_conf",
		WaitingSignal:             "Configuration complete; ready for start up",
		MappedPort:                []string{"9090:80/tcp", "50051:50051/tcp"},
	})
	if err != nil {
		log.Fatalf("failed starting gateway container: %s", err)
	}
	d.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (d *HTTPDeleteUserITSuite) TearDownSuite() {
	if err := d.userPgContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating user postgres container: %s", err)
	}
	if err := d.authPgContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating auth postgres container: %s", err)
	}
	if err := d.redisContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
	if err := d.minioContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating minio container: %s", err)
	}
	if err := d.natsContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating nats container: %s", err)
	}
	if err := d.authContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating auth service container: %s", err)
	}
	if err := d.userServiceContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := d.fileServiceContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := d.notificationServiceContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating notification service container: %s", err)
	}
	if err := d.mailHogContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating mailhog container: %s", err)
	}
	if err := d.gatewayContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}
	log.Println("Tear Down integration test suite for HTTPDeleteUserITSuite")

}
func TestHTTPDeletePasswordITSuite(t *testing.T) {
	suite.Run(t, &HTTPDeleteUserITSuite{})
}

func (d *HTTPDeleteUserITSuite) TestDeleteUserIT_Success() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, d.T())

	client := http.Client{}
	response, err := client.Do(request)
	d.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	d.NoError(err)

	d.Equal(http.StatusCreated, response.StatusCode)
	d.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		d.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", d.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	d.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	d.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	d.NoError(err)

	d.Equal(http.StatusOK, verifyResponse.StatusCode)
	d.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, d.T())

	client = http.Client{}
	response, err = client.Do(request)
	d.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	d.Equal(http.StatusOK, response.StatusCode)
	d.NoError(err)
	d.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	d.NoError(err)

	jwt, ok := respData["data"].(string)
	d.True(ok, "expected jwt token in data field")

	// delete user
	reqBody := &bytes.Buffer{}

	encoder := gin.H{
		"password": "password123",
	}
	_ = json.NewEncoder(reqBody).Encode(encoder)

	request, err = http.NewRequest(http.MethodDelete, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Authorization", "Bearer "+jwt)

	d.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	d.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	d.Equal(http.StatusOK, response.StatusCode)
	d.NoError(err)
	d.Contains(string(byteBody), dto.SUCCESS_DELETE_USER)

}
func (d *HTTPDeleteUserITSuite) TestDeleteUserIT_WrongPassword() {

	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, d.T())

	client := http.Client{}
	response, err := client.Do(request)
	d.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	d.NoError(err)

	d.Equal(http.StatusCreated, response.StatusCode)
	d.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		d.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", d.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	d.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	d.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	d.NoError(err)

	d.Equal(http.StatusOK, verifyResponse.StatusCode)
	d.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, d.T())

	client = http.Client{}
	response, err = client.Do(request)
	d.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	d.Equal(http.StatusOK, response.StatusCode)
	d.NoError(err)
	d.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	d.NoError(err)

	jwt, ok := respData["data"].(string)
	d.True(ok, "expected jwt token in data field")

	// delete user
	reqBody := &bytes.Buffer{}

	encoder := gin.H{
		"password": "password123456",
	}
	_ = json.NewEncoder(reqBody).Encode(encoder)

	request, err = http.NewRequest(http.MethodDelete, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Authorization", "Bearer "+jwt)

	d.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	d.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	d.Equal(http.StatusUnauthorized, response.StatusCode)
	d.NoError(err)
	d.Contains(string(byteBody), dto.Err_UNAUTHORIZED_PASSWORD_WRONG.Error())
}

// case when secret key is leaked
// because the app use stateful session, it makes the verify service
// now that there is no key with the token id in redis that should be
// there if you want to make a request.
// so this test is not applicable here and also handling the error that really edge case
// in normal scenario. need to add security testing to catch this

// func (d *HTTPDeleteUserITSuite) TestDeleteUserIT_UserNotFound() {
// 	// login with dummy generated token
// 	token, _, err := helper.GenerateToken("userid-123", 2*time.Hour)
// 	d.NoError(err)
// 	// delete user
// 	reqBody := &bytes.Buffer{}

// 	encoder := gin.H{
// 		"password": "password123456",
// 	}
// 	_ = json.NewEncoder(reqBody).Encode(encoder)

// 	request, err := http.NewRequest(http.MethodDelete, "http://localhost:9090/api/v1/user/", reqBody)
// 	request.Header.Set("Authorization", "Bearer "+token)

// 	d.NoError(err)

// 	client := http.Client{}
// 	response, err := client.Do(request)
// 	d.NoError(err)

// 	byteBody, err := io.ReadAll(response.Body)

// 	d.Equal(http.StatusNotFound, response.StatusCode)
// 	d.NoError(err)
// 	d.Contains(string(byteBody), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
// }
