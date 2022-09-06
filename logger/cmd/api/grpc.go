package main

import (
	"context"
	"logger/data"
	"logger/logs"
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
