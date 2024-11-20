package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"ryg-user-service/conf"
	"ryg-user-service/db"
	"ryg-user-service/gen_proto/user_service"
	"ryg-user-service/rabbit_mq"
	"ryg-user-service/service"
)

func main() {
	cnf := conf.LoadConfig()

	db.ConnectDB(cnf.DB)
	defer db.CloseDB()

	pm := rabbit_mq.NewPublisherManager(cnf.RabbitMQConfig)
	defer pm.Close()

	lis, err := net.Listen("tcp", cnf.GRPCUrl)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	s := service.NewUserService(db.DB, pm.GenericEmailQueuePublisher)
	user_service.RegisterUserServiceServer(grpcServer, s)

	fmt.Printf("User Microservice is running on port %v...", cnf.GRPCUrl)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
