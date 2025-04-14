package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	api "github.com/etesami/ping-pong-go/api"
	pb "github.com/etesami/ping-pong-go/pkg/protoc"

	// "github.com/prometheus/client_golang/prometheus/promhttp"
	// "google.golang.org/genproto/googleapis/api/metric"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func generateRandomBytes(sizeMB float64) ([]byte, error) {
	size := int(sizeMB * 1024 * 1024)
	randomBytes := make([]byte, size)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}

func sendData(client pb.MessageClient, sizeMB float64) error {
	randomBytes, err := generateRandomBytes(sizeMB)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sentTimestamp := time.Now()
	ack, err := client.SendData(ctx, &pb.Data{
		Payload:       randomBytes,
		SentTimestamp: fmt.Sprintf("%d", int(sentTimestamp.UnixMilli())),
	})
	if err != nil {
		return fmt.Errorf("send data not successful: %v", err)
	}
	fileSize := float64(len(randomBytes)) / (1024)
	log.Printf("Sent [%.2f] KB. Ack recevied, status: [%s], Ack size: [%.2f] KB\n", fileSize, ack.Status, float64(len(ack.Payload))/1024)

	return nil
}

func sendDataInit(client pb.MessageClient, fileSize string) {
	if fileSize == "" {
		log.Println("FILE_SIZE environment variable not set. Using default value of 1 MB.")
		fileSize = "1"
	}
	fileSizeInt, err := strconv.ParseFloat(fileSize, 10)
	if err != nil {
		log.Fatalf("Invalid FILE_SIZE value: %v. Using default value of 1MB.", err)
		fileSizeInt = 1
	}
	if err := sendData(client, fileSizeInt); err != nil {
		log.Printf("Error sending data: %v", err)
	}
}

func main() {
	// Target service initialization
	svcTargetAddress := os.Getenv("SVC_ADDR")
	svcTargetPort := os.Getenv("SVC_PORT")

	targetSvc := &api.Service{
		Address: svcTargetAddress,
		Port:    svcTargetPort,
	}

	var conn *grpc.ClientConn
	var client pb.MessageClient

	for {
		var err error
		conn, err = grpc.NewClient(
			targetSvc.Address+":"+targetSvc.Port,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to target service: %v", err)
			continue
		}
		client = pb.NewMessageClient(conn)
		log.Printf("Connected to target service: %s:%s\n", targetSvc.Address, targetSvc.Port)
		break
	}

	fileSize := os.Getenv("FILE_SIZE")
	updateFrequencyStr := os.Getenv("UPDATE_FREQUENCY")

	if updateFrequencyStr == "" {
		log.Println("UPDATE_FREQUENCY environment variable not set. Sending data and exit.")
		sendDataInit(client, fileSize)
		return
	}

	updateFrequency, err := strconv.Atoi(updateFrequencyStr)
	if err != nil {
		log.Fatalf("Error parsing update frequency: %v", err)
	}
	ticker := time.NewTicker(time.Duration(updateFrequency) * time.Second)
	defer ticker.Stop()

	// Send data initially
	sendDataInit(client, fileSize)

	go func(c *pb.MessageClient) {
		for range ticker.C {
			sendDataInit(*c, fileSize)
		}
	}(&client)

	select {}

	// metricAddr := os.Getenv("METRIC_ADDR")
	// metricPort := os.Getenv("METRIC_PORT")
	// http.Handle("/metrics", promhttp.Handler())
	// log.Printf("Starting server on :%s\n", metricPort)
	// http.ListenAndServe(fmt.Sprintf("%s:%s", metricAddr, metricPort), nil)
}
