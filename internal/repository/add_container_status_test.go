package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddContainerStatus_Success(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	var capturedBody *bytes.Buffer

	es.On("Do", mock.Anything, mock.MatchedBy(func(req esapi.IndexRequest) bool {
		buf := new(bytes.Buffer)
		io.Copy(buf, req.Body)
		capturedBody = buf
		return req.Index == "container_status"
	})).Return(&esapi.Response{
		StatusCode: 201,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil)

	err := repo.AddContainerStatus(context.Background(), 1, "running")
	assert.NoError(t, err)
	assert.NotNil(t, capturedBody)

	var payload map[string]interface{}
	err = json.Unmarshal(capturedBody.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), payload["id"])
	assert.Equal(t, "running", payload["status"])

	es.AssertExpectations(t)
}

func TestAddContainerStatus_InvalidStatus(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	err := repo.AddContainerStatus(context.Background(), 1, "invalid_status")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
	es.AssertNotCalled(t, "Do", mock.Anything, mock.Anything)
}

func TestAddContainerStatus_ESRequestError(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	es.On("Do", mock.Anything, mock.Anything).
		Return(&esapi.Response{
			Body: io.NopCloser(strings.NewReader("")),
		}, errors.New("network error"))

	err := repo.AddContainerStatus(context.Background(), 1, "stopped")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "index request failed")
	es.AssertExpectations(t)
}

func TestAddContainerStatus_ESResponseError(t *testing.T) {
	es := new(MockESClient)
	logger := new(MockLogger)
	repo := &containerRepository{es: es, logger: logger}

	es.On("Do", mock.Anything, mock.Anything).
		Return(&esapi.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(`{"error": "internal failure"}`)),
		}, nil)

	err := repo.AddContainerStatus(context.Background(), 1, "running")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "elasticsearch error")
	es.AssertExpectations(t)
}
