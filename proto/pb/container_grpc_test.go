package pb

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Mock client connection interface
type mockClientConn struct {
	mock.Mock
}

func (m *mockClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	ret := m.Called(ctx, method, args, reply)
	return ret.Error(0)
}

func (m *mockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}

func TestGetAllContainers_Success(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewContainerAdmServiceClient(mockConn)

	expectedResp := &ContainerResponse{
		Containers: []*ContainerName{{Id: 1, ContainerName: "demo"}},
	}
	mockConn.On("Invoke", mock.Anything, ContainerAdmService_GetAllContainers_FullMethodName, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		reply := args.Get(3).(*ContainerResponse)
		cloned := proto.Clone(expectedResp).(*ContainerResponse)
		proto.Merge(reply, cloned)
	}).Return(nil)

	resp, err := client.GetAllContainers(context.Background(), &EmptyRequest{})
	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockConn.AssertExpectations(t)
}

func TestGetAllContainers_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewContainerAdmServiceClient(mockConn)

	mockConn.On("Invoke", mock.Anything, ContainerAdmService_GetAllContainers_FullMethodName, mock.Anything, mock.Anything).Return(errors.New("grpc error"))

	resp, err := client.GetAllContainers(context.Background(), &EmptyRequest{})
	assert.Nil(t, resp)
	assert.EqualError(t, err, "grpc error")
}

func TestGetContainerUptimeDuration_Success(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewContainerAdmServiceClient(mockConn)

	expectedResp := &GetContainerUptimeDurationResponse{
		NumContainers: 5,
	}

	mockConn.On("Invoke",
		mock.Anything,
		ContainerAdmService_GetContainerUptimeDuration_FullMethodName,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Run(func(args mock.Arguments) {
		out := args.Get(3).(*GetContainerUptimeDurationResponse)
		cloned := proto.Clone(expectedResp).(*GetContainerUptimeDurationResponse)
		proto.Merge(out, cloned)
	}).Return(nil)

	req := &GetContainerInfomationRequest{
		StartTime: 1000,
		EndTime:   2000,
	}
	resp, err := client.GetContainerUptimeDuration(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockConn.AssertExpectations(t)
}

func TestGetContainerUptimeDuration_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewContainerAdmServiceClient(mockConn)

	mockConn.On("Invoke",
		mock.Anything,
		ContainerAdmService_GetContainerUptimeDuration_FullMethodName,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("invoke failed"))

	req := &GetContainerInfomationRequest{
		StartTime: 1000,
		EndTime:   2000,
	}
	resp, err := client.GetContainerUptimeDuration(context.Background(), req)

	assert.Nil(t, resp)
	assert.EqualError(t, err, "invoke failed")
	mockConn.AssertExpectations(t)
}

type mockServiceRegistrar struct {
	mock.Mock
}

func (m *mockServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	m.Called(desc, impl)
}

func TestRegisterContainerAdmServiceServer(t *testing.T) {
	mockRegistrar := new(mockServiceRegistrar)

	type mockServer struct {
		UnimplementedContainerAdmServiceServer
	}

	srv := &mockServer{}

	mockRegistrar.On("RegisterService", &ContainerAdmService_ServiceDesc, srv).Return()

	RegisterContainerAdmServiceServer(mockRegistrar, srv)

	mockRegistrar.AssertCalled(t, "RegisterService", &ContainerAdmService_ServiceDesc, srv)
	mockRegistrar.AssertExpectations(t)
}

type mockContainerAdmServiceServer struct {
	mock.Mock
	UnimplementedContainerAdmServiceServer
}

func (m *mockContainerAdmServiceServer) GetAllContainers(ctx context.Context, req *EmptyRequest) (*ContainerResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*ContainerResponse), args.Error(1)
}

func Test_ContainerAdmService_GetAllContainers_Handler_NoInterceptor(t *testing.T) {
	mockServer := &mockContainerAdmServiceServer{}
	mockServer.On("GetAllContainers", mock.Anything, mock.Anything).
		Return(&ContainerResponse{}, nil)

	dec := func(v interface{}) error {
		*v.(*EmptyRequest) = EmptyRequest{}
		return nil
	}

	resp, err := _ContainerAdmService_GetAllContainers_Handler(
		mockServer,
		context.Background(),
		dec,
		nil, // No interceptor
	)

	require.NoError(t, err)
	require.IsType(t, &ContainerResponse{}, resp)
	mockServer.AssertCalled(t, "GetAllContainers", mock.Anything, mock.Anything)
}

func Test_ContainerAdmService_GetAllContainers_Handler_WithInterceptor(t *testing.T) {
	mockServer := &mockContainerAdmServiceServer{}
	mockServer.On("GetAllContainers", mock.Anything, mock.Anything).
		Return(&ContainerResponse{}, nil)

	dec := func(v interface{}) error {
		*v.(*EmptyRequest) = EmptyRequest{}
		return nil
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		require.Equal(t, ContainerAdmService_GetAllContainers_FullMethodName, info.FullMethod)
		return handler(ctx, req)
	}

	resp, err := _ContainerAdmService_GetAllContainers_Handler(
		mockServer,
		context.Background(),
		dec,
		interceptor,
	)

	require.NoError(t, err)
	require.IsType(t, &ContainerResponse{}, resp)
	mockServer.AssertCalled(t, "GetAllContainers", mock.Anything, mock.Anything)
}

func TestGetContainerInformation_Success(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewContainerAdmServiceClient(mockConn)

	req := &GetContainerInfomationRequest{}
	mockConn.On("Invoke", mock.Anything, ContainerAdmService_GetContainerInformation_FullMethodName, req, mock.Anything).
		Return(nil)

	out, err := client.GetContainerInformation(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, out)
	mockConn.AssertExpectations(t)
}

func TestUnimplementedContainerAdmServiceServer_GetAllContainers(t *testing.T) {
	srv := UnimplementedContainerAdmServiceServer{}
	_, err := srv.GetAllContainers(context.Background(), &EmptyRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unimplemented, st.Code())
	require.Contains(t, st.Message(), "GetAllContainers not implemented")
}

func TestUnimplementedContainerAdmServiceServer_GetContainerInformation(t *testing.T) {
	srv := UnimplementedContainerAdmServiceServer{}
	_, err := srv.GetContainerInformation(context.Background(), &GetContainerInfomationRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unimplemented, st.Code())
	require.Contains(t, st.Message(), "GetContainerInformation not implemented")
}

func TestUnimplementedContainerAdmServiceServer_GetContainerUptimeDuration(t *testing.T) {
	srv := UnimplementedContainerAdmServiceServer{}
	_, err := srv.GetContainerUptimeDuration(context.Background(), &GetContainerInfomationRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unimplemented, st.Code())
	require.Contains(t, st.Message(), "GetContainerUptimeDuration not implemented")
}

type mockServer struct {
	UnimplementedContainerAdmServiceServer
}

func (m *mockServer) GetContainerInformation(ctx context.Context, req *GetContainerInfomationRequest) (*GetContainerInfomationResponse, error) {
	return &GetContainerInfomationResponse{
	}, nil
}

func TestGetContainerInformationHandler_NoInterceptor(t *testing.T) {
	srv := &mockServer{}
	ctx := context.Background()

	dec := func(v interface{}) error {
		req := v.(*GetContainerInfomationRequest)
		*req = GetContainerInfomationRequest{} // có thể set field nếu cần
		return nil
	}

	resp, err := _ContainerAdmService_GetContainerInformation_Handler(srv, ctx, dec, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGetContainerInformationHandler_WithInterceptor(t *testing.T) {
	srv := &mockServer{}
	ctx := context.Background()

	dec := func(v interface{}) error {
		*v.(*GetContainerInfomationRequest) = GetContainerInfomationRequest{}
		return nil
	}

	// Interceptor sẽ gọi vào handler truyền vào
	interceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		require.Equal(t, ContainerAdmService_GetContainerInformation_FullMethodName, info.FullMethod)
		return handler(ctx, req)
	}

	resp, err := _ContainerAdmService_GetContainerInformation_Handler(srv, ctx, dec, interceptor)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

type mockContainerAdmServer struct {
	UnimplementedContainerAdmServiceServer
}

func (m *mockContainerAdmServer) GetContainerUptimeDuration(ctx context.Context, req *GetContainerInfomationRequest) (*GetContainerUptimeDurationResponse, error) {
	return &GetContainerUptimeDurationResponse{
	}, nil
}


func Test_GetContainerUptimeDuration_Handler_NoInterceptor(t *testing.T) {
	srv := &mockContainerAdmServer{}
	ctx := context.Background()

	dec := func(v interface{}) error {
		*v.(*GetContainerInfomationRequest) = GetContainerInfomationRequest{}
		return nil
	}

	resp, err := _ContainerAdmService_GetContainerUptimeDuration_Handler(srv, ctx, dec, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_GetContainerUptimeDuration_Handler_WithInterceptor(t *testing.T) {
	srv := &mockContainerAdmServer{}
	ctx := context.Background()

	dec := func(v interface{}) error {
		*v.(*GetContainerInfomationRequest) = GetContainerInfomationRequest{}
		return nil
	}

	interceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		require.Equal(t, ContainerAdmService_GetContainerUptimeDuration_FullMethodName, info.FullMethod)
		return handler(ctx, req)
	}

	resp, err := _ContainerAdmService_GetContainerUptimeDuration_Handler(srv, ctx, dec, interceptor)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
