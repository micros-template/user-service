package it

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dropboks/proto-user/pkg/upb"
	_helper "github.com/dropboks/sharedlib/test/helper"
	"github.com/dropboks/user-service/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type GRPCUpdateUserITSuite struct {
	suite.Suite
	ctx context.Context

	network              *testcontainers.DockerNetwork
	userPgContainer      *_helper.PostgresContainer
	authPgContainer      *_helper.PostgresContainer
	redisContainer       *_helper.RedisContainer
	natsContainer        *_helper.NatsContainer
	authContainer        *_helper.AuthServiceContainer
	userServiceContainer *_helper.UserServiceContainer
}

func (c *GRPCUpdateUserITSuite) SetupSuite() {

	log.Println("Setting up integration test suite for GRPCUpdateUserITSuite")
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
	userPgContainer, err := _helper.StartPostgresContainer(c.ctx, c.network.Name, "user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres user container: %s", err)
	}
	c.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(c.ctx, c.network.Name, "auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres auth container: %s", err)
	}
	c.authPgContainer = authPgContainer

	// spawn redis
	rContainer, err := _helper.StartRedisContainer(c.ctx, c.network.Name, viper.GetString("container.redis_version"))
	if err != nil {
		log.Fatalf("failed starting redis container: %s", err)
	}
	c.redisContainer = rContainer

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

	// spawn user service
	uContainer, err := _helper.StartUserServiceContainer(c.ctx, c.network.Name, viper.GetString("container.user_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting user service container: %s", err)
	}
	c.userServiceContainer = uContainer

}

func (c *GRPCUpdateUserITSuite) TearDownSuite() {
	if err := c.userPgContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating user postgres container: %s", err)
	}
	if err := c.authPgContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating auth postgres container: %s", err)
	}
	if err := c.redisContainer.Terminate(c.ctx); err != nil {
		log.Fatalf("error terminating redis container: %s", err)
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
	log.Println("Tear Down integration test suite for GRPCUpdateUserITSuite")

}
func TestGRPCUpdateUserITSuite(t *testing.T) {
	suite.Run(t, &GRPCUpdateUserITSuite{})
}
func (c *GRPCUpdateUserITSuite) TestUpdateUserIT_Success() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())
	conn, err := helper.ConnectGRPC("localhost:50051")
	c.Require().NoError(err, "Failed to connect to gRPC server")
	defer conn.Close()
	userServiceClient := upb.NewUserServiceClient(conn)

	// register
	// need to fill all
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

	status, err := userServiceClient.CreateUser(c.ctx, user)

	c.NoError(err)
	c.NotNil(status)
	c.Equal(true, status.Success)

	us := &upb.User{
		Id:               "123",
		FullName:         "new-full-name",
		Image:            &image,
		Email:            email,
		Password:         "hashed-password",
		Verified:         true,
		TwoFactorEnabled: false,
	}

	up, err := userServiceClient.UpdateUser(c.ctx, us)

	c.NoError(err)
	c.NotNil(up)
	c.Equal(true, status.Success)
}

func (c *GRPCUpdateUserITSuite) TestUpdateUserIT_UserNotFound() {
	email := fmt.Sprintf("test+%d@example.com", time.Now().UnixNano())

	user := &upb.User{
		Id:       "1234",
		FullName: "full-name",
		Email:    email,
		Password: "password123",
	}
	conn, err := helper.ConnectGRPC("localhost:50051")
	c.Require().NoError(err, "Failed to connect to gRPC server")
	defer conn.Close()

	userServiceClient := upb.NewUserServiceClient(conn)
	status, err := userServiceClient.UpdateUser(c.ctx, user)

	c.Error(err)
	c.Nil(status)
}
