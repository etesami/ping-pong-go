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
	pb "github.com/etesami/ping-pong-go/pkg/protoc"

	// "github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedMessageServer
}

func (s Server) SendData(ctx context.Context, recData *pb.Data) (*pb.Ack, error) {
	st := time.Now()
	recTimestamp := st.UnixMilli()
	log.Printf("Received at [%s]: [%d]\n", st.Format("2006-01-02 15:04:05"), len(recData.Payload))

	ack := &pb.Ack{
		Status:                "ok",
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

	// We listen on all interfaces
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", localSvc.Address, localSvc.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Listening on %s:%s\n", localSvc.Address, localSvc.Port)

	grpcServer := grpc.NewServer()
	pb.RegisterMessageServer(grpcServer, Server{})

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
