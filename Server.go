package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"slices"
	"sync"
	"time"

	pb "github.com/r03smus/auction/proto/auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server structure
type server struct {
	pb.AuctionServer
	pb.ReplicaServer
	replicas        []pb.ReplicaClient
	mu              sync.Mutex // Protect shared resources
	highestBid      int64
	highestBidId    int64
	bidders         []int64
	finished        bool
	uniqeidentifier sync.Map // Thread-safe map for unique identifiers
}

// SendUpdate broadcasts updates to replicas
func (s *server) SendUpdate(ctx context.Context, req *pb.Update) (*pb.Ack, error) {
	var wg sync.WaitGroup
	var errOccurred bool

	for _, replica := range s.replicas {
		wg.Add(1)

		go func(replica pb.ReplicaClient) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_, err := replica.Update(ctx, req)
			if err != nil {
				log.Printf("Replica Update Error: %v", err)
				errOccurred = true
			}
		}(replica)
	}
	wg.Wait()

	if errOccurred {
		return nil, fmt.Errorf("one or more replicas failed to update")
	}
	return &pb.Ack{Ok: true}, nil
}

// Connect handles new replica connections
func (s *server) Connect(ctx context.Context, cr *pb.ConnectRequest) (*pb.Ack, error) {
	log.Printf("Connecting to replica: %s\n", cr.ReplicaAdress)
	fmt.Printf("Connecting to replica: %s\n", cr.ReplicaAdress)

	conn, err := grpc.NewClient(cr.ReplicaAdress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to replica: %v", err)
	}

	client := pb.NewReplicaClient(conn)

	s.replicas = append(s.replicas, client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Update(ctx, &pb.Update{
		Uniqeidentifier: 0,
		Response:        nil,
	})

	return &pb.Ack{Ok: true}, nil
}

// Result retrieves or calculates the result for a unique identifier
func (s *server) Result(ctx context.Context, req *pb.Request) (*pb.Response, error) {

	log.Printf("Result was requested\n")

	if value, ok := s.uniqeidentifier.Load(req.UniqeIdentifier); ok {
		log.Printf("Uniqe identifier found: %s\n", value.(string))
		fmt.Printf("Uniqe identifier found: %s\n", value.(string))
		return value.(*pb.Response), nil
	}

	var state int64
	if s.finished {
		state = 1
	}

	n := &pb.Response{
		State:      state,
		HighestBid: s.highestBid,
	}
	s.uniqeidentifier.Store(req.UniqeIdentifier, n)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.SendUpdate(ctx, &pb.Update{Uniqeidentifier: req.UniqeIdentifier, Response: n})
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Bid handles bid requests
func (s *server) Bid(ctx context.Context, req *pb.BidMessage) (*pb.Response, error) {

	log.Printf("ID: %d bid %d\n", req.Id, req.Bid)
	fmt.Printf("ID: %d bid %d\n", req.Id, req.Bid)

	if value, ok := s.uniqeidentifier.Load(req.UniqeIdentifier); ok {
		return value.(*pb.Response), nil
	}

	var state int64

	s.mu.Lock()

	if s.finished {
		state = 1
		return &pb.Response{State: state}, nil
	}

	if !slices.Contains(s.bidders, req.Id) {
		s.bidders = append(s.bidders, req.Id)
	}
	if s.highestBid < req.Bid {
		s.highestBid = req.Bid
		s.highestBidId = req.Id
	} else {
		state = 1
	}

	response := &pb.Response{State: state}
	s.uniqeidentifier.Store(req.UniqeIdentifier, response)
	s.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.SendUpdate(ctx, &pb.Update{Uniqeidentifier: req.UniqeIdentifier, Response: response})
	if err != nil {
		return nil, err
	}

	return response, nil
}

// endAuction marks the auction as finished after the specified duration
func (s *server) endAuction(duration int64) {
	time.AfterFunc(time.Second*time.Duration(duration), func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.finished = true
		log.Println("Auction Closed!")
		fmt.Println("Auction Closed!")
	})
}

// newServer creates a new server instance
func newServer(duration int64) *server {
	s := &server{}
	s.endAuction(duration)
	return s
}

func newReplicaServer(s *server) *server {
	return s
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	duration := flag.Int("d", 120, "Duration for auction, in seconds (default: 120)")
	flag.Parse()

	lis, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	server := newServer(int64(*duration))
	pb.RegisterAuctionServer(s, server)
	pb.RegisterReplicaServer(s, newReplicaServer(server))

	log.Printf("Server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
