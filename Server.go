package main

import (
	"context"
	pb "github.com/r03smus/auction/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

type server struct {
	pb.AuctionServer
	mu             sync.Mutex // Protects lamportTime and clients map
	highest_bid    int64
	highest_bid_id int64
	bidders        []int
}

func main() {
	lis, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuctionServer(s, newServer())

	log.Println("Server started on port 42069")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func newServer() *server {
	return &server{}
}

func (s *server) Result(ctx context.Context, req *pb.Request) (*pb.RequestRespone, error) {

	return &pb.RequestRespone{
		HighestBid: s.highest_bid,
	}, nil
}

func (s *server) Bid(ctx context.Context, req *pb.BidMessage) (*pb.Ack, error) {

	s.mu.Lock()
	s.highest_bid = req.Bid
	s.highest_bid_id = req.Id
	s.mu.Unlock()

	return &pb.Ack{
		State: 0, // 0 == Success, 1 == Fail, 2 == Exception.
	}, nil
}
