package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/r03smus/auction/proto/auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"slices"
	"sync"
	"time"
)

func (s *server) SendUpdate(ctx context.Context, req *pb.Update) (*pb.Ack, error) {
	for _, replica := range s.replicas {
		go func(replica pb.ReplicaClient) {
			_, err := replica.Update(ctx, req)
			if err != nil {
				log.Fatalf("Replica Update Error: %v", err)
			}
		}(replica)
	}
	return &pb.Ack{Ok: true}, nil
}

func (s *server) Connect(ctx context.Context, cr *pb.ConnectRequest) (*pb.Ack, error) {
	fmt.Println(cr.ReplicaAdress)

	conn, err := grpc.NewClient(cr.ReplicaAdress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("NewClient failed: %v", err)
	}
	client := pb.NewReplicaClient(conn)

	s.replicas = append(s.replicas, client)

	return &pb.Ack{Ok: true}, nil
}

type server struct {
	pb.AuctionServer
	replicas        []pb.ReplicaClient
	mu              sync.Mutex // Protects lamportTime and clients map
	highestBid      int64
	highestBidId    int64
	bidders         []int64
	finished        bool
	uniqeidentifier map[int64]pb.Response
}

func main() {
	duration := flag.Int("d", 120, "Duration for auction, in seconds, Default 120 seconds")

	lis, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuctionServer(s, newServer(int64(*duration)))

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func newServer(duration int64) *server {

	s := &server{}
	s.endAuction(duration)
	return s
}

func (s *server) endAuction(durantion int64) {
	time.AfterFunc(time.Second*time.Duration(durantion), func() {
		s.finished = true
		fmt.Println("Auction Closed!")
	})
}

func (s *server) Result(ctx context.Context, req *pb.Request) (*pb.Response, error) {

	response, ok := s.uniqeidentifier[req.UniqeIdentifier]
	if ok {
		return &response, nil
	}
	var state int64 = 0
	if s.finished {
		state = 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	n := &pb.Response{
		State: state,
	}
	s.uniqeidentifier[req.UniqeIdentifier] = *n
	_, err := s.SendUpdate(ctx, &pb.Update{Uniqeidentifier: req.UniqeIdentifier, Response: n})
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (s *server) Bid(ctx context.Context, req *pb.BidMessage) (*pb.Response, error) {

	var state int64 = 0

	if s.finished {
		state = 1
		return &pb.Response{
			State: state, // 0 == Success, 1 == Fail, 2 == Exception (?).
		}, nil
	}

	s.mu.Lock()
	if !slices.Contains(s.bidders, req.Id) {
		s.bidders = append(s.bidders, req.Id)
	}

	if s.highestBid <= req.Bid {
		s.highestBid = req.Bid
		s.highestBidId = req.Id
	} else {
		state = 1
	}
	s.mu.Unlock()

	reponse := &pb.Response{
		State: state,
	}

	s.uniqeidentifier[req.UniqeIdentifier] = *reponse

	return reponse, nil
}
