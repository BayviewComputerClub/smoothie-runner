package main

import (
	"context"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/cache"
	"github.com/BayviewComputerClub/smoothie-runner/judging"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"github.com/BayviewComputerClub/smoothie-runner/util"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
)

var (
	smoothieRunner *grpc.Server
)

// GRPC route implementations

type SmoothieRunnerAPI struct{}

func (runner *SmoothieRunnerAPI) Health(ctx context.Context, empty *pb.Empty) (*pb.ServiceHealth, error) {
	return &pb.ServiceHealth{
		NumOfTasksToBeDone: uint64(*shared.TasksToBeDone),
		NumOfTasksInQueue:  uint64(*shared.TasksInQueue),
		NumOfWorkers:       uint64(shared.MAX_THREADS),
	}, nil
}

func (runner *SmoothieRunnerAPI) GetProblemTestDataHash(ctx context.Context, p *pb.ProblemTestDataHashRequest) (*pb.ProblemTestDataHashResponse, error) {
	return &pb.ProblemTestDataHashResponse{Hash: cache.GetHash(p.ProblemId)}, nil
}

// upload cache chunks for test data
func (runner *SmoothieRunnerAPI) UploadProblemTestData(stream pb.SmoothieRunnerAPI_UploadProblemTestDataServer) error {
	first := false
	for {
		req, err := stream.Recv()
		if err != nil {
			err = stream.SendAndClose(&pb.UploadTestDataResponse{Error: err.Error()})
			return err // send nil by default
		}

		if !first {
			util.Info("Test data being uploaded for " + req.ProblemId + " of hash " + req.TestDataHash + ".")
			first = true
		}

		// add byte chunk to temp cache
		// this is not thread safe, but should be fine since it only adds when in a loop
		cache.AddByteChunk(req.ProblemId, req.DataChunk)

		// add to cache when finished
		if req.FinishedUploading {
			err := cache.AddToCacheFromChunks(req.ProblemId, req.TestDataHash)
			if err != nil {
				err = stream.SendAndClose(&pb.UploadTestDataResponse{Error: err.Error()})
				return err // send nil by default
			}

			util.Info("Test data finished uploading for " + req.ProblemId + " of hash " + req.TestDataHash + ".")
			err = stream.SendAndClose(&pb.UploadTestDataResponse{Error: ""})
			return err // send nil by default
		}
	}
}

func (runner *SmoothieRunnerAPI) TestSolution(stream pb.SmoothieRunnerAPI_TestSolutionServer) error {
	// receive initial request with data
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	util.Info("Received request to judge " + req.Problem.ProblemId + " in " + req.Solution.Language + ".")

	// check if problem test data is cached
	if !cache.Match(req.Problem.ProblemId, req.Problem.TestDataHash) {
		err := stream.Send(&pb.TestSolutionResponse{
			TestCaseResult: &pb.TestCaseResult{
				BatchNumber: 0,
				CaseNumber:  0,
				Result:      "",
				ResultInfo:  "",
				Time:        0,
				MemUsage:    0,
			},
			CompletedTesting:   false,
			CompileError:       "",
			TestDataNeedUpload: true,
		})
		util.Info("Test data needs to be uploaded for request to judge " + req.Problem.ProblemId + " in " + req.Solution.Language + ".")
		return err
	}

	stat := make(chan shared.JudgeStatus)              // judging status channel
	streamReceive := make(chan pb.TestSolutionRequest) // stream status

	// whether or not the judge task has been cancelled (so that judging process exists)
	isCancelled := false
	defer func() { isCancelled = true }()

	// add judging request to worker queue
	judging.AddToQueue(judging.JudgeJob{Req: req, Res: stat, Cancelled: &isCancelled})

	// listen for judging stream input in goroutine
	go func() {
		for {
			if isCancelled {
				break
			}
			d, err := stream.Recv()
			if err == io.EOF {
				streamReceive <- pb.TestSolutionRequest{CancelTesting: false,}
				break
			}
			if err != nil {
				util.Warn("stream: " + err.Error())
				streamReceive <- pb.TestSolutionRequest{CancelTesting: false,}
				break
			}

			streamReceive <- *d
		}
	}()

	// combining judging stream and grpc status streams simultaneously
	for {
		select {
		case s := <-stat: // if status update

			if !s.Res.CompletedTesting {
				shared.Debug(fmt.Sprintf("Result for %v-%v: %v %v", s.Res.TestCaseResult.BatchNumber, s.Res.TestCaseResult.CaseNumber, s.Res.TestCaseResult.Result, s.Res.TestCaseResult.ResultInfo))
			}

			err = stream.Send(&s.Res)
			if err != nil {
				util.Warn(err.Error())
				isCancelled = true
				return err
			}

			// if completed judging, leave
			if s.Res.CompletedTesting {
				return nil
			}

		case d := <-streamReceive: // if judging stream has output
			if d.CancelTesting {
				isCancelled = true
				break
			}
		}
	}
}

// rpc server start

func startApiServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", shared.PORT))
	if err != nil {
		util.Fatal("IPC listen error (check if port has been taken):" + err.Error())
	}

	smoothieRunner = grpc.NewServer()

	pb.RegisterSmoothieRunnerAPIServer(smoothieRunner, &SmoothieRunnerAPI{})
	util.Info("Started API on port " + strconv.Itoa(shared.PORT))
	go smoothieRunner.Serve(lis)
}
