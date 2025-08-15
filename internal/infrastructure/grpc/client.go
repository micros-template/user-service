package grpc

import (
	fileProto "github.com/micros-template/proto-file/pkg/fpb"
	"github.com/spf13/viper"
)

func NewFileServiceConnection(manager *GRPCClientManager) fileProto.FileServiceClient {
	fileServiceConnection := manager.GetConnection(viper.GetString("app.grpc.service.file_service"))
	fileServiceClient := fileProto.NewFileServiceClient(fileServiceConnection)
	return fileServiceClient
}
