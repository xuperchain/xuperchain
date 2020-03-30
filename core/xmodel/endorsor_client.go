package xmodel

import (
	"errors"

	"google.golang.org/grpc"
)

// NewEndorsorConn return EndorsorClient
func NewEndorsorConn(addr string) (*grpc.ClientConn, error) {
	conn := &grpc.ClientConn{}
	options := append([]grpc.DialOption{}, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, options...)
	if err != nil {
		return nil, errors.New("New grpcs conn error")
	}
	return conn, nil
}
