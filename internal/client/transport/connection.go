package transport

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	initOnce   sync.Once
	closeOnce  sync.Once
	globalConn *grpc.ClientConn
)

func InitConnection(conf *viper.Viper) error {
	var err error

	initOnce.Do(func() {
		globalConn, err = NewConnection(conf)
	})

	return err
}

func GetConnection() *grpc.ClientConn {
	return globalConn
}

func CloseConnection() error {
	var err error

	closeOnce.Do(func() {
		if globalConn != nil {
			err = globalConn.Close()
			globalConn = nil
		}
	})

	return err
}

func NewConnection(conf *viper.Viper) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if conf.GetBool("server.insecure") {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(nil)
	}

	conn, err := grpc.NewClient(
		conf.GetString("server.address"),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("cant create a grpc client connection: %w", err)
	}

	return conn, nil
}
