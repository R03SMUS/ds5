package main

import (
	"bufio"
	"flag"
	"fmt"
	pb "github.com/r03smus/auction/proto/auction"
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
	client     pb.AuctionClient
	Id         int64
	operations int64
}

func (c *Client) PlaceBid(amount int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, _ := c.client.Bid(ctx, &pb.BidMessage{Id: c.Id, Bid: amount, UniqeIdentifier: c.Id*1000 + c.operations})

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
	resultresponse, _ := c.client.Result(ctx, &pb.Request{UniqeIdentifier: c.Id*1000 + c.operations})

	var state string

	if resultresponse.State == 0 {
		state = "open"
	} else {
		state = "closed"
	}

	fmt.Printf("Auction is %s - Highest bid: %d\n", state, resultresponse.HighestBid)
}

func newClient(id int64) *Client {
	conn, err := grpc.NewClient("localhost:42069", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	client := pb.NewAuctionClient(conn)
	fmt.Printf("Started Client: %d\n", id)
	return &Client{
		client: client,
		Id:     id,
	}
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	id := flag.Int("id", 1, "Id")
	flag.Parse()
	client := newClient(int64(*id))
	for {
		client.operations += 1
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
		} else if strings.Contains(input, "/exit") {
			break
		} else {
			fmt.Println("no command")
		}

	}
}
