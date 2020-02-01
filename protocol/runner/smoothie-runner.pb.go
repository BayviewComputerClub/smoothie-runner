// Code generated by protoc-gen-go. DO NOT EDIT.
// source: smoothie-runner.proto

/*
Package smoothie_runner is a generated protocol buffer package.

It is generated from these files:
	smoothie-runner.proto

It has these top-level messages:
	Empty
	ServiceHealth
	TestSolutionRequest
	Solution
	Problem
	ProblemGrader
	ProblemBatch
	ProblemBatchCase
	TestSolutionResponse
	TestCaseResult
*/
package runner

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ServiceHealth struct {
	NumOfTasksToBeDone uint64 `protobuf:"varint,1,opt,name=numOfTasksToBeDone" json:"numOfTasksToBeDone,omitempty"`
	NumOfTasksInQueue  uint64 `protobuf:"varint,2,opt,name=numOfTasksInQueue" json:"numOfTasksInQueue,omitempty"`
	NumOfWorkers       uint64 `protobuf:"varint,3,opt,name=numOfWorkers" json:"numOfWorkers,omitempty"`
}

func (m *ServiceHealth) Reset()                    { *m = ServiceHealth{} }
func (m *ServiceHealth) String() string            { return proto.CompactTextString(m) }
func (*ServiceHealth) ProtoMessage()               {}
func (*ServiceHealth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ServiceHealth) GetNumOfTasksToBeDone() uint64 {
	if m != nil {
		return m.NumOfTasksToBeDone
	}
	return 0
}

func (m *ServiceHealth) GetNumOfTasksInQueue() uint64 {
	if m != nil {
		return m.NumOfTasksInQueue
	}
	return 0
}

func (m *ServiceHealth) GetNumOfWorkers() uint64 {
	if m != nil {
		return m.NumOfWorkers
	}
	return 0
}

type TestSolutionRequest struct {
	Problem               *Problem  `protobuf:"bytes,1,opt,name=problem" json:"problem,omitempty"`
	Solution              *Solution `protobuf:"bytes,2,opt,name=solution" json:"solution,omitempty"`
	TestBatchEvenIfFailed bool      `protobuf:"varint,3,opt,name=testBatchEvenIfFailed" json:"testBatchEvenIfFailed,omitempty"`
	CancelTesting         bool      `protobuf:"varint,4,opt,name=cancelTesting" json:"cancelTesting,omitempty"`
}

func (m *TestSolutionRequest) Reset()                    { *m = TestSolutionRequest{} }
func (m *TestSolutionRequest) String() string            { return proto.CompactTextString(m) }
func (*TestSolutionRequest) ProtoMessage()               {}
func (*TestSolutionRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *TestSolutionRequest) GetProblem() *Problem {
	if m != nil {
		return m.Problem
	}
	return nil
}

func (m *TestSolutionRequest) GetSolution() *Solution {
	if m != nil {
		return m.Solution
	}
	return nil
}

func (m *TestSolutionRequest) GetTestBatchEvenIfFailed() bool {
	if m != nil {
		return m.TestBatchEvenIfFailed
	}
	return false
}

func (m *TestSolutionRequest) GetCancelTesting() bool {
	if m != nil {
		return m.CancelTesting
	}
	return false
}

type Solution struct {
	Language string `protobuf:"bytes,1,opt,name=language" json:"language,omitempty"`
	Code     string `protobuf:"bytes,2,opt,name=code" json:"code,omitempty"`
}

func (m *Solution) Reset()                    { *m = Solution{} }
func (m *Solution) String() string            { return proto.CompactTextString(m) }
func (*Solution) ProtoMessage()               {}
func (*Solution) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Solution) GetLanguage() string {
	if m != nil {
		return m.Language
	}
	return ""
}

func (m *Solution) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

type Problem struct {
	TestBatches       []*ProblemBatch `protobuf:"bytes,1,rep,name=testBatches" json:"testBatches,omitempty"`
	ProblemID         string          `protobuf:"bytes,2,opt,name=problemID" json:"problemID,omitempty"`
	TestCasesHashCode uint64          `protobuf:"varint,3,opt,name=testCasesHashCode" json:"testCasesHashCode,omitempty"`
	Grader            *ProblemGrader  `protobuf:"bytes,4,opt,name=grader" json:"grader,omitempty"`
	TimeLimit         float64         `protobuf:"fixed64,5,opt,name=timeLimit" json:"timeLimit,omitempty"`
	MemLimit          float64         `protobuf:"fixed64,6,opt,name=memLimit" json:"memLimit,omitempty"`
}

func (m *Problem) Reset()                    { *m = Problem{} }
func (m *Problem) String() string            { return proto.CompactTextString(m) }
func (*Problem) ProtoMessage()               {}
func (*Problem) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Problem) GetTestBatches() []*ProblemBatch {
	if m != nil {
		return m.TestBatches
	}
	return nil
}

func (m *Problem) GetProblemID() string {
	if m != nil {
		return m.ProblemID
	}
	return ""
}

func (m *Problem) GetTestCasesHashCode() uint64 {
	if m != nil {
		return m.TestCasesHashCode
	}
	return 0
}

func (m *Problem) GetGrader() *ProblemGrader {
	if m != nil {
		return m.Grader
	}
	return nil
}

func (m *Problem) GetTimeLimit() float64 {
	if m != nil {
		return m.TimeLimit
	}
	return 0
}

func (m *Problem) GetMemLimit() float64 {
	if m != nil {
		return m.MemLimit
	}
	return 0
}

type ProblemGrader struct {
	Type       string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	CustomCode string `protobuf:"bytes,2,opt,name=customCode" json:"customCode,omitempty"`
}

func (m *ProblemGrader) Reset()                    { *m = ProblemGrader{} }
func (m *ProblemGrader) String() string            { return proto.CompactTextString(m) }
func (*ProblemGrader) ProtoMessage()               {}
func (*ProblemGrader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ProblemGrader) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *ProblemGrader) GetCustomCode() string {
	if m != nil {
		return m.CustomCode
	}
	return ""
}

type ProblemBatch struct {
	Cases []*ProblemBatchCase `protobuf:"bytes,1,rep,name=cases" json:"cases,omitempty"`
}

func (m *ProblemBatch) Reset()                    { *m = ProblemBatch{} }
func (m *ProblemBatch) String() string            { return proto.CompactTextString(m) }
func (*ProblemBatch) ProtoMessage()               {}
func (*ProblemBatch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *ProblemBatch) GetCases() []*ProblemBatchCase {
	if m != nil {
		return m.Cases
	}
	return nil
}

type ProblemBatchCase struct {
	Input          string `protobuf:"bytes,1,opt,name=input" json:"input,omitempty"`
	ExpectedAnswer string `protobuf:"bytes,2,opt,name=expectedAnswer" json:"expectedAnswer,omitempty"`
}

func (m *ProblemBatchCase) Reset()                    { *m = ProblemBatchCase{} }
func (m *ProblemBatchCase) String() string            { return proto.CompactTextString(m) }
func (*ProblemBatchCase) ProtoMessage()               {}
func (*ProblemBatchCase) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *ProblemBatchCase) GetInput() string {
	if m != nil {
		return m.Input
	}
	return ""
}

func (m *ProblemBatchCase) GetExpectedAnswer() string {
	if m != nil {
		return m.ExpectedAnswer
	}
	return ""
}

type TestSolutionResponse struct {
	TestCaseResult   *TestCaseResult `protobuf:"bytes,1,opt,name=testCaseResult" json:"testCaseResult,omitempty"`
	CompletedTesting bool            `protobuf:"varint,2,opt,name=completedTesting" json:"completedTesting,omitempty"`
	CompileError     string          `protobuf:"bytes,3,opt,name=compileError" json:"compileError,omitempty"`
}

func (m *TestSolutionResponse) Reset()                    { *m = TestSolutionResponse{} }
func (m *TestSolutionResponse) String() string            { return proto.CompactTextString(m) }
func (*TestSolutionResponse) ProtoMessage()               {}
func (*TestSolutionResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *TestSolutionResponse) GetTestCaseResult() *TestCaseResult {
	if m != nil {
		return m.TestCaseResult
	}
	return nil
}

func (m *TestSolutionResponse) GetCompletedTesting() bool {
	if m != nil {
		return m.CompletedTesting
	}
	return false
}

func (m *TestSolutionResponse) GetCompileError() string {
	if m != nil {
		return m.CompileError
	}
	return ""
}

type TestCaseResult struct {
	BatchNumber uint64  `protobuf:"varint,1,opt,name=batchNumber" json:"batchNumber,omitempty"`
	CaseNumber  uint64  `protobuf:"varint,2,opt,name=caseNumber" json:"caseNumber,omitempty"`
	Result      string  `protobuf:"bytes,3,opt,name=result" json:"result,omitempty"`
	ResultInfo  string  `protobuf:"bytes,4,opt,name=resultInfo" json:"resultInfo,omitempty"`
	Time        float64 `protobuf:"fixed64,5,opt,name=time" json:"time,omitempty"`
	MemUsage    float64 `protobuf:"fixed64,6,opt,name=memUsage" json:"memUsage,omitempty"`
}

func (m *TestCaseResult) Reset()                    { *m = TestCaseResult{} }
func (m *TestCaseResult) String() string            { return proto.CompactTextString(m) }
func (*TestCaseResult) ProtoMessage()               {}
func (*TestCaseResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *TestCaseResult) GetBatchNumber() uint64 {
	if m != nil {
		return m.BatchNumber
	}
	return 0
}

func (m *TestCaseResult) GetCaseNumber() uint64 {
	if m != nil {
		return m.CaseNumber
	}
	return 0
}

func (m *TestCaseResult) GetResult() string {
	if m != nil {
		return m.Result
	}
	return ""
}

func (m *TestCaseResult) GetResultInfo() string {
	if m != nil {
		return m.ResultInfo
	}
	return ""
}

func (m *TestCaseResult) GetTime() float64 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *TestCaseResult) GetMemUsage() float64 {
	if m != nil {
		return m.MemUsage
	}
	return 0
}

func init() {
	proto.RegisterType((*Empty)(nil), "Empty")
	proto.RegisterType((*ServiceHealth)(nil), "ServiceHealth")
	proto.RegisterType((*TestSolutionRequest)(nil), "TestSolutionRequest")
	proto.RegisterType((*Solution)(nil), "Solution")
	proto.RegisterType((*Problem)(nil), "Problem")
	proto.RegisterType((*ProblemGrader)(nil), "ProblemGrader")
	proto.RegisterType((*ProblemBatch)(nil), "ProblemBatch")
	proto.RegisterType((*ProblemBatchCase)(nil), "ProblemBatchCase")
	proto.RegisterType((*TestSolutionResponse)(nil), "TestSolutionResponse")
	proto.RegisterType((*TestCaseResult)(nil), "TestCaseResult")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SmoothieRunnerAPI service

type SmoothieRunnerAPIClient interface {
	TestSolution(ctx context.Context, opts ...grpc.CallOption) (SmoothieRunnerAPI_TestSolutionClient, error)
	Health(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServiceHealth, error)
}

type smoothieRunnerAPIClient struct {
	cc *grpc.ClientConn
}

func NewSmoothieRunnerAPIClient(cc *grpc.ClientConn) SmoothieRunnerAPIClient {
	return &smoothieRunnerAPIClient{cc}
}

func (c *smoothieRunnerAPIClient) TestSolution(ctx context.Context, opts ...grpc.CallOption) (SmoothieRunnerAPI_TestSolutionClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SmoothieRunnerAPI_serviceDesc.Streams[0], c.cc, "/SmoothieRunnerAPI/TestSolution", opts...)
	if err != nil {
		return nil, err
	}
	x := &smoothieRunnerAPITestSolutionClient{stream}
	return x, nil
}

type SmoothieRunnerAPI_TestSolutionClient interface {
	Send(*TestSolutionRequest) error
	Recv() (*TestSolutionResponse, error)
	grpc.ClientStream
}

type smoothieRunnerAPITestSolutionClient struct {
	grpc.ClientStream
}

func (x *smoothieRunnerAPITestSolutionClient) Send(m *TestSolutionRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *smoothieRunnerAPITestSolutionClient) Recv() (*TestSolutionResponse, error) {
	m := new(TestSolutionResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *smoothieRunnerAPIClient) Health(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*ServiceHealth, error) {
	out := new(ServiceHealth)
	err := grpc.Invoke(ctx, "/SmoothieRunnerAPI/Health", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SmoothieRunnerAPI service

type SmoothieRunnerAPIServer interface {
	TestSolution(SmoothieRunnerAPI_TestSolutionServer) error
	Health(context.Context, *Empty) (*ServiceHealth, error)
}

func RegisterSmoothieRunnerAPIServer(s *grpc.Server, srv SmoothieRunnerAPIServer) {
	s.RegisterService(&_SmoothieRunnerAPI_serviceDesc, srv)
}

func _SmoothieRunnerAPI_TestSolution_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SmoothieRunnerAPIServer).TestSolution(&smoothieRunnerAPITestSolutionServer{stream})
}

type SmoothieRunnerAPI_TestSolutionServer interface {
	Send(*TestSolutionResponse) error
	Recv() (*TestSolutionRequest, error)
	grpc.ServerStream
}

type smoothieRunnerAPITestSolutionServer struct {
	grpc.ServerStream
}

func (x *smoothieRunnerAPITestSolutionServer) Send(m *TestSolutionResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *smoothieRunnerAPITestSolutionServer) Recv() (*TestSolutionRequest, error) {
	m := new(TestSolutionRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SmoothieRunnerAPI_Health_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SmoothieRunnerAPIServer).Health(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/SmoothieRunnerAPI/Health",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SmoothieRunnerAPIServer).Health(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _SmoothieRunnerAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "SmoothieRunnerAPI",
	HandlerType: (*SmoothieRunnerAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Health",
			Handler:    _SmoothieRunnerAPI_Health_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TestSolution",
			Handler:       _SmoothieRunnerAPI_TestSolution_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "smoothie-runner.proto",
}

func init() { proto.RegisterFile("smoothie-runner.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 658 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x54, 0xd1, 0x6e, 0x13, 0x3b,
	0x10, 0xed, 0xb6, 0x4d, 0x9a, 0x4c, 0x9a, 0xdc, 0xd6, 0xb7, 0xbd, 0x8a, 0xaa, 0x2b, 0x14, 0x59,
	0x50, 0x22, 0x04, 0x0b, 0x0a, 0x48, 0x95, 0x78, 0x6b, 0xd3, 0x42, 0x23, 0x21, 0x28, 0x6e, 0x10,
	0xcf, 0x9b, 0xcd, 0x34, 0x59, 0x75, 0xd7, 0x5e, 0x6c, 0x6f, 0xa1, 0xfc, 0x05, 0x5f, 0xc0, 0x7f,
	0x20, 0x3e, 0x88, 0xcf, 0x40, 0xf6, 0x7a, 0x93, 0xdd, 0x36, 0x6f, 0x9e, 0x73, 0xc6, 0x9e, 0x99,
	0x33, 0xe3, 0x81, 0x7d, 0x95, 0x08, 0xa1, 0xe7, 0x11, 0x3e, 0x93, 0x19, 0xe7, 0x28, 0xfd, 0x54,
	0x0a, 0x2d, 0xe8, 0x16, 0xd4, 0xce, 0x92, 0x54, 0xdf, 0xd2, 0x1f, 0x1e, 0xb4, 0x2f, 0x51, 0xde,
	0x44, 0x21, 0x9e, 0x63, 0x10, 0xeb, 0x39, 0xf1, 0x81, 0xf0, 0x2c, 0xf9, 0x70, 0x35, 0x0e, 0xd4,
	0xb5, 0x1a, 0x8b, 0x13, 0x3c, 0x15, 0x1c, 0xbb, 0x5e, 0xcf, 0xeb, 0x6f, 0xb2, 0x15, 0x0c, 0x79,
	0x0a, 0xbb, 0x4b, 0x74, 0xc4, 0x3f, 0x66, 0x98, 0x61, 0x77, 0xdd, 0xba, 0xdf, 0x27, 0x08, 0x85,
	0x6d, 0x0b, 0x7e, 0x16, 0xf2, 0x1a, 0xa5, 0xea, 0x6e, 0x58, 0xc7, 0x0a, 0x46, 0x7f, 0x7b, 0xf0,
	0xef, 0x18, 0x95, 0xbe, 0x14, 0x71, 0xa6, 0x23, 0xc1, 0x19, 0x7e, 0xc9, 0x50, 0x69, 0x42, 0x61,
	0x2b, 0x95, 0x62, 0x12, 0x63, 0x62, 0xd3, 0x69, 0x0d, 0x1a, 0xfe, 0x45, 0x6e, 0xb3, 0x82, 0x20,
	0x8f, 0xa0, 0xa1, 0xdc, 0x35, 0x9b, 0x44, 0x6b, 0xd0, 0xf4, 0x17, 0xef, 0x2c, 0x28, 0xf2, 0x0a,
	0xf6, 0x35, 0x2a, 0x7d, 0x12, 0xe8, 0x70, 0x7e, 0x76, 0x83, 0x7c, 0x74, 0xf5, 0x26, 0x88, 0x62,
	0x9c, 0xda, 0x7c, 0x1a, 0x6c, 0x35, 0x49, 0x1e, 0x42, 0x3b, 0x0c, 0x78, 0x88, 0xb1, 0xc9, 0x2e,
	0xe2, 0xb3, 0xee, 0xa6, 0xf5, 0xae, 0x82, 0xf4, 0x35, 0x34, 0x8a, 0x88, 0xe4, 0x00, 0x1a, 0x71,
	0xc0, 0x67, 0x59, 0x30, 0xcb, 0x25, 0x6c, 0xb2, 0x85, 0x4d, 0x08, 0x6c, 0x86, 0x62, 0x9a, 0x6b,
	0xd5, 0x64, 0xf6, 0x4c, 0xff, 0x78, 0xb0, 0xe5, 0x6a, 0x22, 0xcf, 0xa1, 0xb5, 0x48, 0x03, 0x55,
	0xd7, 0xeb, 0x6d, 0xf4, 0x5b, 0x83, 0x76, 0x51, 0xb2, 0x85, 0x59, 0xd9, 0x83, 0xfc, 0x0f, 0x4d,
	0x27, 0xc3, 0xe8, 0xd4, 0xbd, 0xba, 0x04, 0x4c, 0x9f, 0x8c, 0xf3, 0x30, 0x50, 0xa8, 0xce, 0x03,
	0x35, 0x1f, 0x9a, 0xd8, 0xb9, 0xfc, 0xf7, 0x09, 0x72, 0x08, 0xf5, 0x99, 0x0c, 0xa6, 0x28, 0x6d,
	0x8d, 0xad, 0x41, 0xa7, 0x88, 0xfb, 0xd6, 0xa2, 0xcc, 0xb1, 0x26, 0xa6, 0x8e, 0x12, 0x7c, 0x17,
	0x25, 0x91, 0xee, 0xd6, 0x7a, 0x5e, 0xdf, 0x63, 0x4b, 0xc0, 0x94, 0x9f, 0x60, 0x92, 0x93, 0x75,
	0x4b, 0x2e, 0x6c, 0x3a, 0x84, 0x76, 0xe5, 0x49, 0xa3, 0x87, 0xbe, 0x4d, 0x0b, 0x9d, 0xec, 0x99,
	0x3c, 0x00, 0x08, 0x33, 0xa5, 0x45, 0x32, 0x5c, 0x2a, 0x55, 0x42, 0xe8, 0x11, 0x6c, 0x97, 0xf5,
	0x20, 0x8f, 0xa1, 0x16, 0x9a, 0x3a, 0x9c, 0x5a, 0xbb, 0x15, 0xb5, 0x4c, 0x85, 0x2c, 0xe7, 0xe9,
	0x05, 0xec, 0xdc, 0xa5, 0xc8, 0x1e, 0xd4, 0x22, 0x9e, 0x66, 0xda, 0x65, 0x90, 0x1b, 0xe4, 0x10,
	0x3a, 0xf8, 0x2d, 0xc5, 0x50, 0xe3, 0xf4, 0x98, 0xab, 0xaf, 0x28, 0x5d, 0x1a, 0x77, 0x50, 0xfa,
	0xd3, 0x83, 0xbd, 0xea, 0xd4, 0xaa, 0x54, 0x70, 0x85, 0xe4, 0x08, 0x3a, 0x85, 0xbe, 0x0c, 0x55,
	0x16, 0x6b, 0x37, 0xbd, 0xff, 0xf8, 0xe3, 0x0a, 0xcc, 0xee, 0xb8, 0x91, 0x27, 0xb0, 0x13, 0x8a,
	0x24, 0x8d, 0x51, 0xe3, 0xb4, 0x98, 0xb8, 0x75, 0x3b, 0x71, 0xf7, 0x70, 0xf3, 0xaf, 0x0c, 0x16,
	0xc5, 0x78, 0x26, 0xa5, 0x90, 0xb6, 0xb1, 0x4d, 0x56, 0xc1, 0xe8, 0x2f, 0x0f, 0x3a, 0xd5, 0x90,
	0xa4, 0x07, 0xad, 0x89, 0xa9, 0xff, 0x7d, 0x96, 0x4c, 0x50, 0xba, 0x5f, 0x5e, 0x86, 0x6c, 0x07,
	0x02, 0x85, 0xce, 0x21, 0xff, 0xd7, 0x25, 0x84, 0xfc, 0x07, 0x75, 0x99, 0x57, 0x95, 0x87, 0x74,
	0x96, 0xb9, 0x97, 0x9f, 0x46, 0xfc, 0x4a, 0xd8, 0x21, 0x6a, 0xb2, 0x12, 0x62, 0xbb, 0x1d, 0x25,
	0xe8, 0x66, 0xc6, 0x9e, 0xdd, 0xb8, 0x7c, 0x52, 0xe6, 0xb7, 0x2c, 0xc7, 0xc5, 0xda, 0x83, 0xef,
	0xb0, 0x7b, 0xe9, 0x56, 0x19, 0xb3, 0x9b, 0xec, 0xf8, 0x62, 0x44, 0x8e, 0x61, 0xbb, 0x2c, 0x39,
	0xd9, 0xf3, 0x57, 0xec, 0x8d, 0x83, 0x7d, 0x7f, 0x55, 0x5f, 0xe8, 0x5a, 0xdf, 0x7b, 0xe1, 0x11,
	0x0a, 0x75, 0xb7, 0xf8, 0xea, 0xbe, 0x5d, 0x89, 0x07, 0x1d, 0xbf, 0xb2, 0x10, 0xe9, 0xda, 0xa4,
	0x6e, 0x97, 0xe6, 0xcb, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x51, 0x76, 0x2b, 0x78, 0x4d, 0x05,
	0x00, 0x00,
}