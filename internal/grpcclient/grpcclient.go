package grpcclient

import (
	"fmt"

	"google.golang.org/grpc"
)

func Conn(hostname string, port int) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", hostname, port), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("could not connect: %w", err)
	}

	return conn, nil
}
