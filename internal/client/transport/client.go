package transport

import (
	"sync"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	pbAuth "github.com/kuvalkin/gophkeeper/internal/proto/auth/v1"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
)

var (
	initAuthOnce sync.Once
	initSyncOnce sync.Once

	globalAuthClient pbAuth.AuthServiceClient
	globalSyncClient pbSync.SyncServiceClient
)

func GetAuthClient(conf *viper.Viper) (pbAuth.AuthServiceClient, error) {
	var err error

	initAuthOnce.Do(func() {
		err = InitConnection(conf)

		if err == nil {
			globalAuthClient = NewAuthClient(GetConnection())
		}
	})

	return globalAuthClient, err
}

func GetSyncClient(conf *viper.Viper) (pbSync.SyncServiceClient, error) {
	var err error

	initSyncOnce.Do(func() {
		err = InitConnection(conf)

		if err == nil {
			globalSyncClient = NewSyncClient(GetConnection())
		}
	})

	return globalSyncClient, err
}

func NewAuthClient(conn *grpc.ClientConn) pbAuth.AuthServiceClient {
	return pbAuth.NewAuthServiceClient(conn)
}

func NewSyncClient(conn *grpc.ClientConn) pbSync.SyncServiceClient {
	return pbSync.NewSyncServiceClient(conn)
}
