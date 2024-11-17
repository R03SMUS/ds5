package main

import (
	"bufio"
	"fmt"
	pb "github.com/r03smus/auction/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	client pb.AuctionClient
	Id     int64
}

func (c *Client) PlaceBid(amount int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, _ := c.client.Bid(ctx, &pb.BidMessage{Id: c.Id, Bid: amount})

	if response.State == 0 {
		fmt.Println("bid was successful")
	} else if response.State == 1 {
		fmt.Println("bid was not successful")
	} else {
		fmt.Println("some exception")
	}
}

func (c *Client) RequestBid() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resultresponse, _ := c.client.Result(ctx, &pb.Request{})
	fmt.Println("Highest bid:", resultresponse.HighestBid)
}

func newClient(id int64) *Client {
	conn, err := grpc.NewClient("localhost:42069", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	client := pb.NewAuctionClient(conn)
	return &Client{
		client: client,
		Id:     id,
	}
}

func main() {
	client := newClient(0)
	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		if strings.Contains(input, "/bid") {
			list := strings.Split(input, " ")

			item2 := strings.TrimSpace(list[1])

			amount, err := strconv.ParseInt(item2, 10, 64)
			if err != nil {
				fmt.Println("Not a int (int64)", item2)
				continue
			}

			client.PlaceBid(amount)

		} else if strings.Contains(input, "/request") {
			client.RequestBid()
		} else {
			fmt.Println("no command")
		}

	}
	client.RequestBid()
	client.PlaceBid(6200)
	client.RequestBid()

}
