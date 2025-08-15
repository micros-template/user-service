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

type GRPCDeleteUserITSuite struct {
	suite.Suite
	ctx context.Context

	network              *testcontainers.DockerNetwork
	gatewayContainer     *_helper.GatewayContainer
	userPgContainer      *_helper.PostgresContainer
	authPgContainer      *_helper.PostgresContainer
	redisContainer       *_helper.RedisContainer
	minioContainer       *_helper.MinioContainer
	natsContainer        *_helper.NatsContainer
	authContainer        *_helper.AuthServiceContainer
	userServiceContainer *_helper.UserServiceContainer
	fileServiceContainer *_helper.FileServiceContainer
}

func (d *GRPCDeleteUserITSuite) SetupSuite() {
	log.Println("Setting up integration test suite for GRPCDeleteUserITSuite")
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
	userPgContainer, err := _helper.StartPostgresContainer(d.ctx, d.network.Name, "test_user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	d.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(d.ctx, d.network.Name, "test_auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	d.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartRedisContainer(d.ctx, d.network.Name, viper.GetString("container.redis_version"))
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	d.redisContainer = rContainer

	mContainer, err := _helper.StartMinioContainer(d.ctx, d.network.Name, viper.GetString("container.minio_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	d.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartNatsContainer(d.ctx, d.network.Name, viper.GetString("container.nats_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	d.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(d.ctx, d.network.Name, viper.GetString("container.auth_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting auth service container: %s", err)
	}
	d.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(d.ctx, d.network.Name, viper.GetString("container.file_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting file service container: %s", err)
	}
	d.fileServiceContainer = fContainer
	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(d.ctx, d.network.Name, viper.GetString("container.user_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting user service container: %s", err)
	}
	d.userServiceContainer = uContainer

	gatewayContainer, err := _helper.StartGatewayContainer(d.ctx, d.network.Name, viper.GetString("container.gateway_version"))
	if err != nil {
		log.Fatalf("failed starting gateway container: %s", err)
	}
	d.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (d *GRPCDeleteUserITSuite) TearDownSuite() {
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
	if err := d.fileServiceContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := d.userServiceContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := d.gatewayContainer.Terminate(d.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for GRPCDeleteUserITSuite")
}
func TestGRPDeleteUserITSuite(t *testing.T) {
	suite.Run(t, &GRPCDeleteUserITSuite{})
}
func (d *GRPCDeleteUserITSuite) TestDeleteUserIT_Success() {
	// create user
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	user := &upb.User{
		Id:       "123",
		FullName: "full-name",
		Email:    email,
		Password: "password123",
	}
	conn, err := helper.ConnectGRPC("localhost:50051")
	d.Require().NoError(err, "Failed to connect to gRPC server")
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	userServiceClient := upb.NewUserServiceClient(conn)
	status, err := userServiceClient.CreateUser(d.ctx, user)

	d.Require().NoError(err)
	d.Require().NotNil(status)
	d.Require().Equal(true, status.Success)

	status2, err := userServiceClient.DeleteUser(d.ctx, &upb.UserId{
		UserId: user.GetId(),
	})

	d.NoError(err)
	d.NotNil(status2)
	d.Equal(true, status2.Success)
}

func (d *GRPCDeleteUserITSuite) TestDeleteUserIT_UserNotFound() {

	conn, err := helper.ConnectGRPC("localhost:50051")
	d.Require().NoError(err, "Failed to connect to gRPC server")
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	userServiceClient := upb.NewUserServiceClient(conn)

	status2, err := userServiceClient.DeleteUser(d.ctx, &upb.UserId{
		UserId: "123",
	})

	d.Error(err)
	d.Nil(status2)
}
