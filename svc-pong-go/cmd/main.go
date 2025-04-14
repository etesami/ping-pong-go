package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	api "github.com/etesami/ping-pong-go/api"
	// metrics "github.com/etesami/ping-pong-go/pkg/metric"
	pb "github.com/etesami/ping-pong-go/pkg/protoc"
	util "github.com/etesami/ping-pong-go/pkg/utils"

	// "github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMessageServer
	ackSize float64
}

// CheckConnection is a simple ping-pong method to respond for the health check
func (s Server) CheckConnection(ctx context.Context, recData *pb.Data) (*pb.Ack, error) {
	t := time.Now()
	ack := &pb.Ack{
		Status:                "pong",
		Payload:               []byte("pong"),
		OriginalSentTimestamp: recData.SentTimestamp,
		ReceivedTimestamp:     fmt.Sprintf("%d", int(t.UnixMilli())),
		AckSentTimestamp:      fmt.Sprintf("%d", int(t.UnixMilli())),
	}
	return ack, nil
}

func (s Server) SendData(ctx context.Context, recData *pb.Data) (*pb.Ack, error) {
	st := time.Now()
	recTimestamp := st.UnixMilli()

	randomBytes, err := util.GenerateRandomBytes(s.ackSize)
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
	log.Printf("Received at [%s]: [%.2f] KB, Sent Ack: [%.2f] KB\n",
		st.Format("2006-01-02 15:04:05"), float64(len(recData.Payload))/1024, float64(len(ack.Payload))/1024)
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
	// log.Printf("Starting server on %s:%s\n", metricAddr, metricPort)
	// http.Handle("/metrics", promhttp.Handler())
	// http.ListenAndServe(fmt.Sprintf("%s:%s", metricAddr, metricPort), nil)
}
