package client

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/stretchr/testify/require"
)

func TestRealESClient_Do_Failure(t *testing.T) {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9999"}, // wrong port to force failure
	}
	es, err := elasticsearch.NewClient(cfg)
	require.NoError(t, err)

	client := &RealESClient{Client: es}
	req := esapi.InfoRequest{} 

	resp, err := client.Do(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)
}
