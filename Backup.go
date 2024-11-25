package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/r03smus/auction/proto/auction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// replica structure
type replica struct {
	pb.ReplicaServer
	uniqeidentifier sync.Map // Thread-safe map for unique identifiers
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	// Parse command-line flag for replica ID
	id := flag.String("id", "50000", "Replica address, e.g., 50000")
	flag.Parse()

	replicaAddress := "localhost:" + *id
	fmt.Printf("Starting replica at %s\n", replicaAddress)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+*id)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterReplicaServer(grpcServer, newReplica())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Connect to primary server
	client := connectToPrimary("localhost:42069")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.Connect(ctx, &pb.ConnectRequest{ReplicaAdress: replicaAddress})
	if err != nil {
		log.Fatalf("Failed to connect to primary server: %v", err)
	}

	fmt.Printf("Replica successfully connected to primary at localhost:42069\n")
	select {} // Block forever
}

// newReplica initializes a new replica instance
func newReplica() *replica {
	return &replica{}
}

// connectToPrimary establishes a connection to the primary server
func connectToPrimary(primaryAddress string) pb.ReplicaClient {
	conn, err := grpc.NewClient(primaryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to primary server: %v", err)
	}
	return pb.NewReplicaClient(conn)
}

// Update handles update requests from the primary server
func (r *replica) Update(ctx context.Context, req *pb.Update) (*pb.Ack, error) {
	fmt.Printf("Received update: UniqueIdentifier=%d\n", req.Uniqeidentifier)

	// Check if the unique identifier already exists
	_, loaded := r.uniqeidentifier.LoadOrStore(req.Uniqeidentifier, req.Response)
	if loaded {
		fmt.Println("Update ignored: Entry already exists")
	} else {
		fmt.Println("Update applied")
	}

	return &pb.Ack{Ok: true}, nil
}
