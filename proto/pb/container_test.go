package pb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestEmptyRequest(t *testing.T) {
	req := &EmptyRequest{}
	req.Reset()
	assert.NotNil(t, req.String())
	assert.NotNil(t, req.ProtoReflect())
}

func TestContainerName(t *testing.T) {
	c := &ContainerName{
		Id:            123,
		ContainerName: "test-container",
	}
	assert.Equal(t, uint64(123), c.GetId())
	assert.Equal(t, "test-container", c.GetContainerName())

	c.Reset()
	assert.NotNil(t, c.String())
	assert.NotNil(t, c.ProtoReflect())
}

func TestContainerResponse(t *testing.T) {
	cr := &ContainerResponse{
		Containers: []*ContainerName{
			{Id: 1, ContainerName: "c1"},
			{Id: 2, ContainerName: "c2"},
		},
	}
	assert.Len(t, cr.GetContainers(), 2)

	cr.Reset()
	assert.NotNil(t, cr.String())
	assert.NotNil(t, cr.ProtoReflect())
}

func TestGetContainerInfomationRequest(t *testing.T) {
	req := &GetContainerInfomationRequest{
		StartTime: 1000,
		EndTime:   2000,
	}
	assert.Equal(t, int64(1000), req.GetStartTime())
	assert.Equal(t, int64(2000), req.GetEndTime())

	req.Reset()
	assert.NotNil(t, req.String())
	assert.NotNil(t, req.ProtoReflect())
}

func TestGetContainerInfomationResponse(t *testing.T) {
	resp := &GetContainerInfomationResponse{
		NumContainers:        10,
		NumRunningContainers: 6,
		NumStoppedContainers: 4,
		MeanUptimeRatio:      0.75,
	}

	assert.Equal(t, int64(10), resp.GetNumContainers())
	assert.Equal(t, int64(6), resp.GetNumRunningContainers())
	assert.Equal(t, int64(4), resp.GetNumStoppedContainers())
	assert.Equal(t, float32(0.75), resp.GetMeanUptimeRatio())

	resp.Reset()
	assert.NotNil(t, resp.String())
	assert.NotNil(t, resp.ProtoReflect())
}

func TestGetContainerUptimeDurationResponse(t *testing.T) {
	resp := &GetContainerUptimeDurationResponse{
		NumContainers:        5,
		NumRunningContainers: 3,
		NumStoppedContainers: 2,
		UptimeDetails: &ContainerUptimeDetails{
			TotalUptime: 10000,
			PerContainerUptime: map[string]int64{
				"c1": 6000,
				"c2": 4000,
			},
		},
	}

	assert.Equal(t, int64(5), resp.GetNumContainers())
	assert.Equal(t, int64(3), resp.GetNumRunningContainers())
	assert.Equal(t, int64(2), resp.GetNumStoppedContainers())
	assert.NotNil(t, resp.GetUptimeDetails())

	resp.Reset()
	assert.NotNil(t, resp.String())
	assert.NotNil(t, resp.ProtoReflect())
}

func TestContainerUptimeDetails(t *testing.T) {
	details := &ContainerUptimeDetails{
		TotalUptime: 12345,
		PerContainerUptime: map[string]int64{
			"a": 1000,
			"b": 5000,
		},
	}

	assert.Equal(t, int64(12345), details.GetTotalUptime())
	assert.Equal(t, int64(1000), details.GetPerContainerUptime()["a"])

	details.Reset()
	assert.NotNil(t, details.String())
	assert.NotNil(t, details.ProtoReflect())
}

func TestProtoMarshaling(t *testing.T) {
	original := &ContainerName{
		Id:            99,
		ContainerName: "example",
	}

	data, err := proto.Marshal(original)
	assert.NoError(t, err)

	var decoded ContainerName
	err = proto.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, original.GetId(), decoded.GetId())
	assert.Equal(t, original.GetContainerName(), decoded.GetContainerName())
}


