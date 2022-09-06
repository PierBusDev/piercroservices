package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"logger/data"
	"logger/logs"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()
	//write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := logs.LogResponse{
			Result: "Failed",
		}
		return &res, err
	}

	//return the response
	res := logs.LogResponse{
		Result: "Successfully logged!",
	}
	return &res, nil
}

//gRPCListen starts the gRPC server
func (c *Config) gRPCListen() {
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("[gRPCListen]failed to listen for grpc calls: %v", err)
	}

	grpcServer := grpc.NewServer()
	logs.RegisterLogServiceServer(grpcServer, &LogServer{
		Models: c.Models,
	})

	log.Println("gRPC server listening on port: ", grpcPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[gRPCListen]failed to serve grpc: %v", err)
	}
}
