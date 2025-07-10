package infrastructure

import (
	"fmt"
	"thanhnt208/container-adm-service/config"

	"github.com/elastic/go-elasticsearch/v8"
)

type Elasticsearch struct {
	client *elasticsearch.Client
	cfg    *config.Config
}

var NewElasticsearch = func(cfg *config.Config) IElasticsearch {
	return &Elasticsearch{
		cfg: cfg,
	}
}

func (e *Elasticsearch) ConnectElasticsearch() (*elasticsearch.Client, error) {
	if e.client != nil {
		return e.client, nil
	}

	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{e.cfg.EsAddr},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	e.client = es
	return e.client, nil
}

func (e *Elasticsearch) Close() error {
	return nil
}
