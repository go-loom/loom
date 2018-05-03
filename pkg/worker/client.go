package worker

import (
	"github.com/go-loom/loom/pkg/log"
	"github.com/go-loom/loom/pkg/rpc/pb"

	kitlog "github.com/go-kit/kit/log"

	"context"
	"net/http"
)

type Client struct {
	twirpClient pb.Loom
	logger      kitlog.Logger
}

func NewClientWithHttpClient(url string, c *http.Client) *Client {
	twirpClient := pb.NewLoomProtobufClient(url, c)
	client := &Client{
		twirpClient: twirpClient,
		logger:      log.With(log.Logger),
	}

	return client
}

func NewClient(url string) *Client {
	return NewClientWithHttpClient(url, &http.Client{})
}

func (c *Client) SubscribeJob(ctx context.Context, req *pb.SubscribeJobRequest) (res *pb.SubscribeJobResponse, err error) {
	res, err = c.twirpClient.SubscribeJob(ctx, req)
	return
}

func (c *Client) ReportJob(ctx context.Context, req *pb.ReportJobRequest) (res *pb.ReportJobResponse, err error) {
	res, err = c.twirpClient.ReportJob(ctx, req)
	return
}

func (c *Client) ReportJobDone(ctx context.Context, req *pb.ReportJobDoneRequest) (res *pb.ReportJobDoneResponse, err error) {
	res, err = c.twirpClient.ReportJobDone(ctx, req)
	return
}
