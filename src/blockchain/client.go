package blockchain

import (
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var BLOCKCHAIN_GRPC_ADDRESS = os.Getenv("BLOCKCHAIN_GRPC_ADDRESS")

func NewBlockchainClient() (TransactionServiceClient, *grpc.ClientConn) {
	conn, err := grpc.NewClient(BLOCKCHAIN_GRPC_ADDRESS, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	return NewTransactionServiceClient(conn), conn
}
