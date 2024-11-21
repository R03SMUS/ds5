package main

import (
	"context"
	"flag"
	pb "github.com/r03smus/auction/proto/auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"time"
)

type replica struct {
	pb.AuctionServer
	uniqeidentifier map[int64]pb.Response
}

func main() {
	id := flag.String("id", ":50000", "replica address fx: :50000")

	replicaAdress := "localhost:" + *id

	lis, err := net.Listen("tcp", *id)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuctionServer(s, newReplica())
	s.Serve(lis)

	client := connect("localhost:42069")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client.Connect(ctx, &pb.ConnectRequest{ReplicaAdress: replicaAdress})
}

func newReplica() pb.AuctionServer {
	return &replica{}
}
func connect(primary string) pb.ReplicaClient {
	conn, err := grpc.NewClient(primary, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	return pb.NewReplicaClient(conn)
}

func (s *replica) Update(ctx context.Context, req *pb.Update) (*pb.Ack, error) {
	_, ok := s.uniqeidentifier[req.Uniqeidentifier]
	if !ok {
		s.uniqeidentifier[req.Uniqeidentifier] = *req.Response
	}
	return &pb.Ack{Ok: true}, nil
}
