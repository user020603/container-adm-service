package repository

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetContainerUptimeRatio_Success(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	mockResponse := `{
		"aggregations": {
			"avg_ratio": {
				"value": 0.85
			}
		}
	}`

	es.On("Do", mock.Anything, mock.MatchedBy(func(req esapi.SearchRequest) bool {
		return len(req.Index) == 1 && req.Index[0] == "container_status"
	})).Return(&esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(mockResponse)),
	}, nil)

	ratio, err := repo.GetContainerUptimeRatio(context.Background(), start, end)
	assert.NoError(t, err)
	assert.InDelta(t, 0.85, ratio, 0.0001)

	es.AssertExpectations(t)
}

func TestGetContainerUptimeRatio_StartAfterEnd(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now()
	end := start.Add(-10 * time.Minute) // invalid

	ratio, err := repo.GetContainerUptimeRatio(context.Background(), start, end)
	assert.Error(t, err)
	assert.Equal(t, float64(0), ratio)
	assert.Contains(t, err.Error(), "start time cannot be after end time")
}

func TestGetContainerUptimeRatio_ESResponseError(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	es.On("Do", mock.Anything, mock.Anything).Return(&esapi.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("internal error")),
	}, nil)

	ratio, err := repo.GetContainerUptimeRatio(context.Background(), start, end)
	assert.Error(t, err)
	assert.Equal(t, float64(0), ratio)
	assert.Contains(t, err.Error(), "search error")
}

func TestGetContainerUptimeRatio_DecodeError(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	es.On("Do", mock.Anything, mock.Anything).Return(&esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("invalid json")),
	}, nil)

	ratio, err := repo.GetContainerUptimeRatio(context.Background(), start, end)
	assert.Error(t, err)
	assert.Equal(t, float64(0), ratio)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetContainerUptimeDuration_Success(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	responseBody := `
	{
		"aggregations": {
			"containers": {
				"buckets": [
					{
						"key": 1,
						"running_count": { "value": 10 }
					},
					{
						"key": 2,
						"running_count": { "value": 5 }
					}
				]
			}
		}
	}`

	es.On("Do", mock.Anything, mock.Anything).Return(&esapi.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}, nil)

	res, err := repo.GetContainerUptimeDuration(context.Background(), start, end)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 15*time.Minute, res.TotalUptime)
	assert.Equal(t, 10*time.Minute, res.PerContainerUptime["1"])
	assert.Equal(t, 5*time.Minute, res.PerContainerUptime["2"])

	es.AssertExpectations(t)
}

func TestGetContainerUptimeDuration_InvalidRange(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now()
	end := start.Add(-1 * time.Hour) // invalid

	res, err := repo.GetContainerUptimeDuration(context.Background(), start, end)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "start time cannot be after end time")
}

func TestGetContainerUptimeDuration_ESResponseError(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now()

	body := `{"error": "bad request"}`
	es.On("Do", mock.Anything, mock.Anything).Return(&esapi.Response{
		StatusCode: 400,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil)

	res, err := repo.GetContainerUptimeDuration(context.Background(), start, end)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "search error")
	es.AssertExpectations(t)
}
