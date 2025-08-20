package it

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/micros-template/user-service/internal/domain/dto"
	"github.com/micros-template/user-service/test/helper"

	_helper "github.com/micros-template/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPUpdateUserITSuite struct {
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

func (u *HTTPUpdateUserITSuite) SetupSuite() {
	log.Println("Setting up integration test suite for HTTPUpdateUserITSuite")
	u.ctx = context.Background()

	viper.SetConfigName("config.test")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config")
	}
	// spawn sharedNetwork
	u.network = _helper.StartNetwork(u.ctx)

	// spawn user db
	userPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 u.ctx,
		SharedNetwork:           u.network.Name,
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
	u.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 u.ctx,
		SharedNetwork:           u.network.Name,
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
	u.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartCacheContainer(_helper.CacheParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
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
	u.redisContainer = rContainer

	mContainer, err := _helper.StartStorageContainer(_helper.StorageParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
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
	u.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartMessageQueueContainer(_helper.MessageQueueParameterOption{
		Context:            u.ctx,
		SharedNetwork:      u.network.Name,
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
	u.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(_helper.AuthServiceParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
		ImageName:     viper.GetString("container.auth_service_image"),
		ContainerName: "test_auth_service",
		WaitingSignal: "HTTP Server Starting in port",
		Cmd:           []string{"/auth_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting auth service container: %s", err)
	}
	u.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(_helper.FileServiceParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
		ImageName:     viper.GetString("container.file_service_image"),
		ContainerName: "test_file_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/file_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting file service container: %s", err)
	}
	u.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(_helper.UserServiceParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
		ImageName:     viper.GetString("container.user_service_image"),
		ContainerName: "test_user_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/user_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting user service container: %s", err)
	}
	u.userServiceContainer = uContainer

	noContainer, err := _helper.StartNotificationServiceContainer(_helper.NotificationServiceParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
		ImageName:     viper.GetString("container.notification_service_image"),
		ContainerName: "test_notification_service",
		WaitingSignal: "subscriber for notification is running",
		Cmd:           []string{"/notification_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting notification service container: %s", err)
	}
	u.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailContainer(_helper.MailParameterOption{
		Context:       u.ctx,
		SharedNetwork: u.network.Name,
		ImageName:     viper.GetString("container.mailhog_image"),
		ContainerName: "mailhog",
		WaitingSignal: "1025/tcp",
		MappedPort:    []string{"1025:1025/tcp", "8025:8025/tcp"},
	})
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	u.mailHogContainer = mailContainer

	gatewayContainer, err := _helper.StartGatewayContainer(_helper.GatewayParameterOption{
		Context:                   u.ctx,
		SharedNetwork:             u.network.Name,
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
	u.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (u *HTTPUpdateUserITSuite) TearDownSuite() {
	if err := u.userPgContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating user postgres container: %s", err)
	}
	if err := u.authPgContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating auth postgres container: %s", err)
	}
	if err := u.redisContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
	}
	if err := u.minioContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating minio container: %s", err)
	}
	if err := u.natsContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating nats container: %s", err)
	}
	if err := u.authContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating auth service container: %s", err)
	}
	if err := u.userServiceContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := u.fileServiceContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := u.notificationServiceContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating notification service container: %s", err)
	}
	if err := u.mailHogContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating mailhog container: %s", err)
	}
	if err := u.gatewayContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for HTTPUpdateUserITSuite")
}
func TestHTTPUpdateUserITSuite(t *testing.T) {
	suite.Run(t, &HTTPUpdateUserITSuite{})
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_Success() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, u.T())

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	u.NoError(err)

	u.Equal(http.StatusCreated, response.StatusCode)
	u.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		u.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", u.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	u.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	u.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	u.NoError(err)

	u.Equal(http.StatusOK, verifyResponse.StatusCode)
	u.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, u.T())

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusOK, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	u.NoError(err)

	jwt, ok := respData["data"].(string)
	u.True(ok, "expected jwt token in data field")

	// update user
	reqBody := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	if err := formWriter.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+jwt)
	u.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusOK, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "success update profile data")
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_MissingUserID() {

	reqBody := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	if err := formWriter.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}

	request, err := http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	u.NoError(err)

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	u.Equal(http.StatusUnauthorized, response.StatusCode)
	u.NoError(err)
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_MissingBody() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, u.T())

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	u.NoError(err)

	u.Equal(http.StatusCreated, response.StatusCode)
	u.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		u.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", u.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	u.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	u.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	u.NoError(err)

	u.Equal(http.StatusOK, verifyResponse.StatusCode)
	u.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, u.T())

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusOK, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	u.NoError(err)

	jwt, ok := respData["data"].(string)
	u.True(ok, "expected jwt token in data field")

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", nil)
	request.Header.Set("Authorization", "Bearer "+jwt)

	u.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusBadRequest, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "invalid input")
}

// not applicable due to checking in the previous layer and case (register, login, and verification middleware)

// func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_UserNotFound() {

// 	reqBody := &bytes.Buffer{}
// 	formWriter := multipart.NewWriter(reqBody)
// 	_ = formWriter.WriteField("full_name", "test-full-name")
// 	formWriter.Close()

// 	request, err := http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", reqBody)
// 	request.Header.Set("Content-Type", formWriter.FormDataContentType())

// 	u.NoError(err)
// 	request.Header.Set("User-Data", `{"user_id":"12345"}`)

// 	client := http.Client{}
// 	response, err := client.Do(request)
// 	u.NoError(err)

// 	byteBody, err := io.ReadAll(response.Body)

// 	u.Equal(http.StatusNotFound, response.StatusCode)
// 	u.NoError(err)
// 	u.Contains(string(byteBody), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
// }

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_ImageWrongExtension() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, u.T())

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	u.NoError(err)

	u.Equal(http.StatusCreated, response.StatusCode)
	u.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		u.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", u.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	u.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	u.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	u.NoError(err)

	u.Equal(http.StatusOK, verifyResponse.StatusCode)
	u.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, u.T())

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusOK, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	u.NoError(err)

	jwt, ok := respData["data"].(string)
	u.True(ok, "expected jwt token in data field")

	// update user
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	fileWriter, _ := formWriter.CreateFormFile("image", "test.webp")
	_, err = fileWriter.Write([]byte("fake image data"))
	if err != nil {
		log.Fatal("failed to create image data")
	}
	if err := formWriter.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+jwt)

	u.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusBadRequest, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), dto.Err_BAD_REQUEST_WRONG_EXTENSION.Error())
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_ImageLimitSizeExceeded() {
	// register
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	request := helper.Register(email, u.T())

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)
	u.NoError(err)

	u.Equal(http.StatusCreated, response.StatusCode)
	u.Contains(string(byteBody), "Register Success. Check your email for verification.")
	if err := response.Body.Close(); err != nil {
		u.T().Errorf("error closing response body: %v", err)
	}
	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:9090/api/v1/auth/verify-email\?userid=[^&]+&token=[^"']+`
	link := helper.RetrieveDataFromEmail(email, regex, "mail", u.T())

	verifyRequest, err := http.NewRequest(http.MethodGet, link, nil)
	u.NoError(err)

	verifyResponse, err := client.Do(verifyRequest)
	u.NoError(err)

	verifyBody, err := io.ReadAll(verifyResponse.Body)
	u.NoError(err)

	u.Equal(http.StatusOK, verifyResponse.StatusCode)
	u.Contains(string(verifyBody), "Verification Success")

	time.Sleep(time.Second) //give a time for auth_db update the user

	// login
	request = helper.Login(email, u.T())

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusOK, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "Login Success")

	var respData map[string]interface{}
	err = json.Unmarshal(byteBody, &respData)
	u.NoError(err)

	jwt, ok := respData["data"].(string)
	u.True(ok, "expected jwt token in data field")

	// update user
	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	fileWriter, _ := formWriter.CreateFormFile("image", "test.png")
	largeData := make([]byte, 8*1024*1024)
	_, err = fileWriter.Write(largeData)
	u.NoError(err)
	if err != nil {
		log.Fatal("failed to create image data")
	}
	if err := formWriter.Close(); err != nil {
		log.Fatal("failed to close form writer")
	}
	request, err = http.NewRequest(http.MethodPatch, "http://localhost:9090/api/v1/user/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+jwt)

	u.NoError(err)

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusBadRequest, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED.Error())
}
