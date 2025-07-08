package infrastructure

import (
	"testing"
	"thanhnt208/container-adm-service/config"

	"github.com/stretchr/testify/assert"
)

func mockConfig() *config.Config {
	return &config.Config{
		EsAddr: "http://localhost:9200",
	}
}

func TestNewElasticsearch(t *testing.T) {
	cfg := mockConfig()

	es := NewElasticsearch(cfg)

	assert.NotNil(t, es)
	assert.Nil(t, es.(*Elasticsearch).client)
	assert.Equal(t, cfg, es.(*Elasticsearch).cfg)
}

func TestElasticsearch_ConnectElasticsearch_Success(t *testing.T) {
	es := &Elasticsearch{
		cfg: mockConfig(),
	}

	client, err := es.ConnectElasticsearch()
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, es.client, client)
}

func TestElasticsearch_ConnectElasticsearch_Memoized(t *testing.T) {
	es := &Elasticsearch{
		cfg: mockConfig(),
	}

	first, err := es.ConnectElasticsearch()
	assert.NoError(t, err)

	second, err := es.ConnectElasticsearch()
	assert.NoError(t, err)

	assert.Same(t, first, second)
}

func TestElasticsearch_Close(t *testing.T) {
	es := &Elasticsearch{}
	err := es.Close()
	assert.NoError(t, err)
}
