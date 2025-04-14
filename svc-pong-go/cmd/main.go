package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	api "github.com/etesami/ping-pong-go/api"
	pb "github.com/etesami/ping-pong-go/pkg/protoc"

	// "github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMessageServer
	ackSize float64
}

func generateRandomBytes(sizeMB float64) ([]byte, error) {
	size := int(sizeMB * 1024 * 1024)
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func (s Server) SendData(ctx context.Context, recData *pb.Data) (*pb.Ack, error) {
	st := time.Now()
	recTimestamp := st.UnixMilli()
	log.Printf("Received at [%s]: [%d]\n", st.Format("2006-01-02 15:04:05"), len(recData.Payload))

	randomBytes, err := generateRandomBytes(s.ackSize)
	if err != nil {
		log.Printf("Error generating random bytes: %v", err)
		return nil, err
	}
	ack := &pb.Ack{
		Status:                "ok",
		Payload:               randomBytes,
		OriginalSentTimestamp: recData.SentTimestamp,
		ReceivedTimestamp:     strconv.Itoa(int(recTimestamp)),
		AckSentTimestamp:      strconv.Itoa(int(time.Now().UnixMilli())),
	}
	return ack, nil
}

func main() {

	// Local service initialization
	svcAddress := os.Getenv("SVC_ADDR")
	svcPort := os.Getenv("SVC_PORT")

	localSvc := &api.Service{
		Address: svcAddress,
		Port:    svcPort,
	}

	fileSize := os.Getenv("FILE_SIZE")
	if fileSize == "" {
		log.Println("FILE_SIZE environment variable not set. Using default value of 1 MB.")
		fileSize = "1"
	}
	fileSizeFloat, err := strconv.ParseFloat(fileSize, 10)
	if err != nil {
		log.Fatalf("Invalid FILE_SIZE value: %v. Using default value of 1MB.", err)
		fileSizeFloat = 1.0
	}

	// We listen on all interfaces
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", localSvc.Address, localSvc.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listening on %s:%s\n", localSvc.Address, localSvc.Port)

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServer(grpcServer, Server{ackSize: fileSizeFloat})

	go func() {
		log.Printf("starting gRPC server on port %s:%s\n", localSvc.Address, localSvc.Port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	select {}

	// metricAddr := os.Getenv("METRIC_ADDR")
	// metricPort := os.Getenv("METRIC_PORT")
	// http.Handle("/metrics", promhttp.Handler())
	// log.Printf("Starting server on :%s\n", metricPort)
	// http.ListenAndServe(fmt.Sprintf("%s:%s", metricAddr, metricPort), nil)
}
