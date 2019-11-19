package main

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/xuperchain/xuperunion/pb"
)

func main() {
	conn, err := grpc.Dial("localhost:6718", grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	client := pb.NewPubsubServiceClient(conn)

	stream, err := client.Subscribe(
		context.Background(), &pb.String{Value: "xuper dpzuVdosQrF2kmzumhVeFQZa1aYcdgFpN "},
	)
	if err != nil {
		fmt.Println(err)
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}

		fmt.Println(reply.GetValue())
	}
}
