package it

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/micros-template/user-service/test/helper"

	"github.com/micros-template/proto-user/pkg/upb"
	_helper "github.com/micros-template/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type GRPCCreateUserITSuite struct {
	suite.Suite
	ctx context.Context

	network              *testcontainers.DockerNetwork
	gatewayContainer     *_helper.GatewayContainer
	userPgContainer      *_helper.SQLContainer
	authPgContainer      *_helper.SQLContainer
	redisContainer       *_helper.CacheContainer
	minioContainer       *_helper.StorageContainer
	natsContainer        *_helper.MessageQueueContainer
	authContainer        *_helper.AuthServiceContainer
	userServiceContainer *_helper.UserServiceContainer
	fileServiceContainer *_helper.FileServiceContainer
}

func (c *GRPCCreateUserITSuite) SetupSuite() {
	log.Println("Setting up integration test suite for GRPCCreateUserITSuite")
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
	userPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 c.ctx,
		SharedNetwork:           c.network.Name,
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
	c.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartSQLContainer(_helper.SQLParameterOption{
		Context:                 c.ctx,
		SharedNetwork:           c.network.Name,
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
	c.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartCacheContainer(_helper.CacheParameterOption{
		Context:       c.ctx,
		SharedNetwork: c.network.Name,
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
	c.redisContainer = rContainer

	mContainer, err := _helper.StartStorageContainer(_helper.StorageParameterOption{
		Context:       c.ctx,
		SharedNetwork: c.network.Name,
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
	c.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartMessageQueueContainer(_helper.MessageQueueParameterOption{
		Context:            c.ctx,
		SharedNetwork:      c.network.Name,
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
	c.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(_helper.AuthServiceParameterOption{
		Context:       c.ctx,
		SharedNetwork: c.network.Name,
		ImageName:     viper.GetString("container.auth_service_image"),
		ContainerName: "test_auth_service",
		WaitingSignal: "HTTP Server Starting in port",
		Cmd:           []string{"/auth_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting auth service container: %s", err)
	}
	c.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(_helper.FileServiceParameterOption{
		Context:       c.ctx,
		SharedNetwork: c.network.Name,
		ImageName:     viper.GetString("container.file_service_image"),
		ContainerName: "test_file_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/file_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting file service container: %s", err)
	}
	c.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(_helper.UserServiceParameterOption{
		Context:       c.ctx,
		SharedNetwork: c.network.Name,
		ImageName:     viper.GetString("container.user_service_image"),
		ContainerName: "test_user_service",
		WaitingSignal: "gRPC server running in port",
		Cmd:           []string{"/user_service"},
		Env:           map[string]string{"ENV": "test"},
	})
	if err != nil {
		log.Fatalf("failed starting user service container: %s", err)
	}
	c.userServiceContainer = uContainer

	gatewayContainer, err := _helper.StartGatewayContainer(_helper.GatewayParameterOption{
		Context:                   c.ctx,
		SharedNetwork:             c.network.Name,
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
	c.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (c *GRPCCreateUserITSuite) TearDownSuite() {
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
	if err := c.fileServiceContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := c.userServiceContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := c.gatewayContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for GRPCCreateUserITSuite")
}
func TestGRPCreateUserITSuite(t *testing.T) {
	suite.Run(t, &GRPCCreateUserITSuite{})
}
func (c *GRPCCreateUserITSuite) TestCreateUserIT_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	user := &upb.User{
		Id:       "123",
		FullName: "full-name",
		Email:    email,
		Password: "password123",
	}
	conn, err := helper.ConnectGRPC("localhost:50051")
	c.Require().NoError(err, "Failed to connect to gRPC server")
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	userServiceClient := upb.NewUserServiceClient(conn)
	status, err := userServiceClient.CreateUser(c.ctx, user)

	c.NoError(err)
	c.NotNil(status)
	c.Equal(true, status.Success)
}
