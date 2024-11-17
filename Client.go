package main

import (
	"fmt"
	pb "github.com/r03smus/auction/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	conn, err := grpc.NewClient("localhost:42069", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	client := pb.NewAuctionClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, _ := client.Bid(ctx, &pb.BidMessage{Id: 0, Bid: 200})

	if response.State == 0 {
		fmt.Println("bid was successful")
	} else if response.State == 1 {
		fmt.Println("bid was not successful")
	} else {
		fmt.Println("some exception")
	}

	resultresponse, _ := client.Result(ctx, &pb.Request{})

	fmt.Println("Highest bid:", resultresponse.HighestBid)

}
