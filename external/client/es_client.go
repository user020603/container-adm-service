package client

import (
	"context"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchClient interface {
	Do(ctx context.Context, req esapi.Request) (*esapi.Response, error)
}

type RealESClient struct {
	Client *elasticsearch.Client
}

func (r *RealESClient) Do(ctx context.Context, req esapi.Request) (*esapi.Response, error) {
	return req.Do(ctx, r.Client)
}
