package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/r03smus/auction/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"slices"
	"sync"
	"time"
)

type server struct {
	pb.AuctionServer
	mu             sync.Mutex // Protects lamportTime and clients map
	highest_bid    int64
	highest_bid_id int64
	bidders        []int64
	finished       bool
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

func (s *server) Result(ctx context.Context, req *pb.Request) (*pb.RequestRespone, error) {
	return &pb.RequestRespone{
		HighestBid: s.highest_bid,
	}, nil
}

func (s *server) Bid(ctx context.Context, req *pb.BidMessage) (*pb.Ack, error) {

	if s.finished {
		return &pb.Ack{
			State: 1, // 0 == Success, 1 == Fail, 2 == Exception (?).
		}, nil
	}

	var state int64 = 0
	s.mu.Lock()
	if !slices.Contains(s.bidders, req.Id) {
		s.bidders = append(s.bidders, req.Id)
	}

	if s.highest_bid <= req.Bid {
		s.highest_bid = req.Bid
		s.highest_bid_id = req.Id
	} else {
		state = 1
	}
	s.mu.Unlock()

	return &pb.Ack{
		State: state, // 0 == Success, 1 == Fail, 2 == Exception (?).
	}, nil
}
