package grpc

import (
	fileProto "10.1.20.130/dropping/proto-file/pkg/fpb"
	"github.com/spf13/viper"
)

func NewFileServiceConnection(manager *GRPCClientManager) fileProto.FileServiceClient {
	fileServiceConnection := manager.GetConnection(viper.GetString("app.grpc.service.file_service"))
	fileServiceClient := fileProto.NewFileServiceClient(fileServiceConnection)
	return fileServiceClient
}
