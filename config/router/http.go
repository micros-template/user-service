package router

import (
	"os"
	"strings"
	"time"

	"10.1.20.130/dropping/log-management/pkg"
	"github.com/dropboks/sharedlib/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewHTTP(js jetstream.JetStream, zerolog zerolog.Logger) *gin.Engine {
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	streamName := viper.GetString("jetstream.log.stream.name")
	streamDesc := viper.GetString("jetstream.log.stream.description")
	globalSubject := viper.GetString("jetstream.log.subject.global")
	subjectPrefix := viper.GetString("jetstream.log.subject.prefix")

	logEmitter := pkg.NewLogEmitter(js, zerolog, streamName, streamDesc, globalSubject, subjectPrefix)
	r.Use(middleware.AccessLogger(logEmitter, "user_service", zerolog))

	allowOrigins := viper.GetString("server.cors.allow_origins")
	allowMethods := viper.GetString("server.cors.allow_methods")
	allowHeaders := viper.GetString("server.cors.allow_headers")
	exposeHeaders := viper.GetString("server.cors.expose_headers")
	allowCredentials := viper.GetBool("server.cors.allow_credential")
	maxAge := viper.GetInt("server.cors.max_age")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(allowOrigins, ","),
		AllowMethods:     strings.Split(allowMethods, ","),
		AllowHeaders:     strings.Split(allowHeaders, ","),
		ExposeHeaders:    strings.Split(exposeHeaders, ","),
		AllowCredentials: allowCredentials,
		MaxAge:           time.Duration(maxAge),
	}))

	return r
}
