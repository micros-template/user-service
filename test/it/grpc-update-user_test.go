package it

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"10.1.20.130/dropping/user-service/test/helper"
	"github.com/dropboks/proto-user/pkg/upb"
	_helper "github.com/dropboks/sharedlib/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type GRPCUpdateUserITSuite struct {
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

func (u *GRPCUpdateUserITSuite) SetupSuite() {

	log.Println("Setting up integration test suite for GRPCUpdateUserITSuite")
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
	userPgContainer, err := _helper.StartPostgresContainer(u.ctx, u.network.Name, "test_user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres user container: %s", err)
	}
	u.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(u.ctx, u.network.Name, "test_auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres auth container: %s", err)
	}
	u.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartRedisContainer(u.ctx, u.network.Name, viper.GetString("container.redis_version"))
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	u.redisContainer = rContainer

	mContainer, err := _helper.StartMinioContainer(u.ctx, u.network.Name, viper.GetString("container.minio_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	u.minioContainer = mContainer

	// spawn nats
	nContainer, err := _helper.StartNatsContainer(u.ctx, u.network.Name, viper.GetString("container.nats_version"))
	if err != nil {
		log.Fatalf("failed starting minio container: %s", err)
	}
	u.natsContainer = nContainer

	aContainer, err := _helper.StartAuthServiceContainer(u.ctx, u.network.Name, viper.GetString("container.auth_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting auth service container: %s", err)
	}
	u.authContainer = aContainer

	fContainer, err := _helper.StartFileServiceContainer(u.ctx, u.network.Name, viper.GetString("container.file_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting file service container: %s", err)
	}
	u.fileServiceContainer = fContainer

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(u.ctx, u.network.Name, viper.GetString("container.user_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting user service container: %s", err)
	}
	u.userServiceContainer = uContainer

	gatewayContainer, err := _helper.StartGatewayContainer(u.ctx, u.network.Name, viper.GetString("container.gateway_version"))
	if err != nil {
		log.Fatalf("failed starting gateway container: %s", err)
	}
	u.gatewayContainer = gatewayContainer
	time.Sleep(time.Second)
}

func (u *GRPCUpdateUserITSuite) TearDownSuite() {
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
	if err := u.fileServiceContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating file service container: %s", err)
	}
	if err := u.userServiceContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating user service container: %s", err)
	}
	if err := u.gatewayContainer.Terminate(u.ctx); err != nil {
		log.Fatalf("error terminating gateway container: %s", err)
	}

	log.Println("Tear Down integration test suite for GRPCUpdateUserITSuite")

}
func TestGRPCUpdateUserITSuite(t *testing.T) {
	suite.Run(t, &GRPCUpdateUserITSuite{})
}
func (u *GRPCUpdateUserITSuite) TestUpdateUserIT_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	conn, err := helper.ConnectGRPC("localhost:50051")
	u.Require().NoError(err, "Failed to connect to gRPC server")
	defer conn.Close()
	userServiceClient := upb.NewUserServiceClient(conn)

	// register
	image := "image"
	user := &upb.User{
		Id:               "123",
		FullName:         "full-name",
		Image:            &image,
		Email:            email,
		Password:         "hashed-password",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	status, err := userServiceClient.CreateUser(u.ctx, user)

	u.NoError(err)
	u.NotNil(status)
	u.Equal(true, status.Success)

	us := &upb.User{
		Id:               "123",
		FullName:         "new-full-name",
		Image:            &image,
		Email:            email,
		Password:         "hashed-password",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	up, err := userServiceClient.UpdateUser(u.ctx, us)

	u.NoError(err)
	u.NotNil(up)
	u.Equal(true, status.Success)
}

func (u *GRPCUpdateUserITSuite) TestUpdateUserIT_UserNotFound() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	user := &upb.User{
		Id:       "1234",
		FullName: "full-name",
		Email:    email,
		Password: "password123",
	}
	conn, err := helper.ConnectGRPC("10.1.20.130:50051")
	u.Require().NoError(err, "Failed to connect to gRPC server")
	defer conn.Close()

	userServiceClient := upb.NewUserServiceClient(conn)
	status, err := userServiceClient.UpdateUser(u.ctx, user)

	u.Error(err)
	u.Nil(status)
}
