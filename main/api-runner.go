package main

import (
	"context"
	"fmt"
	"github.com/BayviewComputerClub/smoothie-runner/judging"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
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

func (runner *SmoothieRunnerAPI) TestSolution(stream pb.SmoothieRunnerAPI_TestSolutionServer) error {
	// receive initial request with data
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	util.Info("Received request to judge " + req.Solution.Problem.ProblemID + " in " + req.Solution.Language + ".")

	stat := make(chan shared.JudgeStatus)
	streamReceive := make(chan pb.TestSolutionRequest)

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

	// combining judging stream and grpc status streams simulaneously
	for {
		select {
		case s := <-stat: // if status update

			if !s.Res.CompletedTesting {
				shared.Debug(fmt.Sprintf("Result for %v-%v: %v %v", s.Res.TestCaseResult.BatchNumber, s.Res.TestCaseResult.CaseNumber, s.Res.TestCaseResult.Result, s.Res.TestCaseResult.ResultInfo))
			}

			err = stream.Send(&s.Res)
			if err != nil {
				util.Warn(err.Error())
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
