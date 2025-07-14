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

	_helper "github.com/dropboks/sharedlib/test/helper"
	"github.com/dropboks/sharedlib/utils"
	"github.com/dropboks/user-service/internal/domain/dto"
	"github.com/dropboks/user-service/test/helper"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type HTTPUpdateUserITSuite struct {
	suite.Suite
	ctx context.Context

	network                      *testcontainers.DockerNetwork
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
	userPgContainer, err := _helper.StartPostgresContainer(u.ctx, u.network.Name, "user_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
	}
	u.userPgContainer = userPgContainer

	// spawn auth db
	authPgContainer, err := _helper.StartPostgresContainer(u.ctx, u.network.Name, "auth_db", viper.GetString("container.postgresql_version"))
	if err != nil {
		log.Fatalf("failed starting postgres container: %s", err)
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

	noContainer, err := _helper.StartNotificationServiceContainer(u.ctx, u.network.Name, viper.GetString("container.notification_service_version"))
	if err != nil {
		log.Println("make sure the image is exist")
		log.Fatalf("failed starting notification service container: %s", err)
	}
	u.notificationServiceContainer = noContainer

	mailContainer, err := _helper.StartMailhogContainer(u.ctx, u.network.Name, viper.GetString("container.mailhog_version"))
	if err != nil {
		log.Fatalf("failed starting mailhog container: %s", err)
	}
	u.mailHogContainer = mailContainer
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
	response.Body.Close()

	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:8081/verify-email\?userid=[^&]+&token=[^"']+`
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

	// verify token
	verifyReq, err := http.NewRequest(http.MethodPost, "http://localhost:8081/verify", nil)
	u.NoError(err)
	verifyReq.Header.Set("Authorization", "Bearer "+jwt)

	verifyResp, err := client.Do(verifyReq)
	u.NoError(err)
	defer verifyResp.Body.Close()

	u.Equal(http.StatusNoContent, verifyResp.StatusCode)
	userDataHeader := verifyResp.Header.Get("User-Data")
	u.NotEmpty(userDataHeader, "User-Data header should not be empty")

	var ud utils.UserData
	err = json.Unmarshal([]byte(userDataHeader), &ud)
	u.NoError(err)
	u.NotEmpty(ud.UserId, "user_id should not be empty")

	// update user

	reqBody := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	formWriter.Close()

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:8082/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())

	u.NoError(err)
	request.Header.Set("User-Data", userDataHeader)

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
	formWriter.Close()

	request, err := http.NewRequest(http.MethodPatch, "http://localhost:8082/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())
	u.NoError(err)

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)

	u.Equal(http.StatusUnauthorized, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_MissingBody() {

	request, err := http.NewRequest(http.MethodPatch, "http://localhost:8082/", nil)

	u.NoError(err)
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)

	u.Equal(http.StatusBadRequest, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), "invalid input")
}

func (u *HTTPUpdateUserITSuite) TestUpdateUserIT_UserNotFound() {

	reqBody := &bytes.Buffer{}
	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	formWriter.Close()

	request, err := http.NewRequest(http.MethodPatch, "http://localhost:8082/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())

	u.NoError(err)
	request.Header.Set("User-Data", `{"user_id":"12345"}`)

	client := http.Client{}
	response, err := client.Do(request)
	u.NoError(err)

	byteBody, err := io.ReadAll(response.Body)

	u.Equal(http.StatusNotFound, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), dto.Err_NOTFOUND_USER_NOT_FOUND.Error())
}

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
	response.Body.Close()

	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:8081/verify-email\?userid=[^&]+&token=[^"']+`
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

	// verify token
	verifyReq, err := http.NewRequest(http.MethodPost, "http://localhost:8081/verify", nil)
	u.NoError(err)
	verifyReq.Header.Set("Authorization", "Bearer "+jwt)

	verifyResp, err := client.Do(verifyReq)
	u.NoError(err)
	defer verifyResp.Body.Close()

	u.Equal(http.StatusNoContent, verifyResp.StatusCode)
	userDataHeader := verifyResp.Header.Get("User-Data")
	u.NotEmpty(userDataHeader, "User-Data header should not be empty")

	var ud utils.UserData
	err = json.Unmarshal([]byte(userDataHeader), &ud)
	u.NoError(err)
	u.NotEmpty(ud.UserId, "user_id should not be empty")

	// update user

	reqBody := &bytes.Buffer{}

	formWriter := multipart.NewWriter(reqBody)
	_ = formWriter.WriteField("full_name", "test-full-name")
	fileWriter, _ := formWriter.CreateFormFile("image", "test.webp")
	_, err = fileWriter.Write([]byte("fake image data"))
	if err != nil {
		log.Fatal("failed to create image data")
	}
	formWriter.Close()

	request, err = http.NewRequest(http.MethodPatch, "http://localhost:8082/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())

	u.NoError(err)
	request.Header.Set("User-Data", userDataHeader)

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
	response.Body.Close()

	time.Sleep(time.Second) //give a time for auth_db update the user

	regex := `http://localhost:8081/verify-email\?userid=[^&]+&token=[^"']+`
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

	// verify token
	verifyReq, err := http.NewRequest(http.MethodPost, "http://localhost:8081/verify", nil)
	u.NoError(err)
	verifyReq.Header.Set("Authorization", "Bearer "+jwt)

	verifyResp, err := client.Do(verifyReq)
	u.NoError(err)
	defer verifyResp.Body.Close()

	u.Equal(http.StatusNoContent, verifyResp.StatusCode)
	userDataHeader := verifyResp.Header.Get("User-Data")
	u.NotEmpty(userDataHeader, "User-Data header should not be empty")

	var ud utils.UserData
	err = json.Unmarshal([]byte(userDataHeader), &ud)
	u.NoError(err)
	u.NotEmpty(ud.UserId, "user_id should not be empty")

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
	formWriter.Close()
	request, err = http.NewRequest(http.MethodPatch, "http://localhost:8082/", reqBody)
	request.Header.Set("Content-Type", formWriter.FormDataContentType())

	u.NoError(err)
	request.Header.Set("User-Data", userDataHeader)

	client = http.Client{}
	response, err = client.Do(request)
	u.NoError(err)

	byteBody, err = io.ReadAll(response.Body)

	u.Equal(http.StatusBadRequest, response.StatusCode)
	u.NoError(err)
	u.Contains(string(byteBody), dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED.Error())
}
