package rpc

import (
	pb "github.com/go-loom/loom/pkg/api/server"

	"context"
)

type Server struct{}

func (s *Server) SubscribeJob(ctx context.Context, workerInfo *pb.WorkerInfo) (job *pb.Job, err error) {
	return nil, nil
}

func (s *Server) ReportJob(ctx context.Context, job *pb.Job) (*pb.EmptyResponse, error) {
	return nil, nil
}

func (s *Server) ReportJobDone(ctx context.Context, job *pb.Job) (*pb.EmptyResponse, error) {
	return nil, nil
}
