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

	"github.com/micros-template/user-service/test/helper"

	_helper "github.com/micros-template/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPGetProfileITSuite struct {
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
	userPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 g.ctx,
		SharedNetwork:           g.network.Name,
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
	g.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 g.ctx,
		SharedNetwork:           g.network.Name,
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
	g.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartCacheContainer(_helper.CacheParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
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
	g.redisContainer = rContainer

	mContainer, err := _helper.StartStorageContainer(_helper.StorageParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
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
	g.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartMessageQueueContainer(_helper.MessageQueueParameterOption{
		Context:            g.ctx,
		SharedNetwork:      g.network.Name,
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
	g.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(_helper.AuthServiceParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
		ImageName:     viper.GetString("container.auth_service_image"),
		ContainerName: "test_auth_service",
		WaitingSignal: "HTTP Server Starting in port",
		Cmd:           []string{"/auth_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting auth service container: %s", err)
	}
	g.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(_helper.FileServiceParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
		ImageName:     viper.GetString("container.file_service_image"),
		ContainerName: "test_file_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/file_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting file service container: %s", err)
	}
	g.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(_helper.UserServiceParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
		ImageName:     viper.GetString("container.user_service_image"),
		ContainerName: "test_user_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/user_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting user service container: %s", err)
	}
	g.userServiceContainer = uContainer

	noContainer, err := _helper.StartNotificationServiceContainer(_helper.NotificationServiceParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
		ImageName:     viper.GetString("container.notification_service_image"),
		ContainerName: "test_notification_service",
		WaitingSignal: "subscriber for notification is running",
		Cmd:           []string{"/notification_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting notification service container: %s", err)
	}
	g.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailContainer(_helper.MailParameterOption{
		Context:       g.ctx,
		SharedNetwork: g.network.Name,
		ImageName:     viper.GetString("container.mailhog_image"),
		ContainerName: "mailhog",
		WaitingSignal: "1025/tcp",
		MappedPort:    []string{"1025:1025/tcp", "8025:8025/tcp"},
	})
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	g.mailHogContainer = mailContainer
	fmt.Println(viper.GetString("script.nginx"))
	gatewayContainer, err := _helper.StartGatewayContainer(_helper.GatewayParameterOption{
		Context:                   g.ctx,
		SharedNetwork:             g.network.Name,
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
