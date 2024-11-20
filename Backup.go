package main

import (
	"context"
	"flag"
	pb "github.com/r03smus/auction/proto/auction"
	"google.golang.org/grpc"
	"log"
	"net"
)

type BackupReplica struct {
	pb.AuctionServer

	uniqeidentifier map[int64]pb.Response
}

func main() {

	id := flag.String("id", "42070", "port value ex \"42070\"")

	lis, err := net.Listen("tcp", ":"+*id)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuctionServer(s, newBackup())

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func newBackup() *BackupReplica {
	return &BackupReplica{}
}

func (br *BackupReplica) Update(ctx context.Context, req *pb.Update) (*pb.Ack, error) {

	_, ok := br.uniqeidentifier[req.Uniqeidentifier]
	if !ok {
		br.uniqeidentifier[req.Uniqeidentifier] = *req.Response
	}

	return &pb.Ack{
		Ok: true,
	}, nil
}
