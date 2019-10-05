package main

import (
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

func (runner *SmoothieRunnerAPI) TestSolution(stream pb.SmoothieRunnerAPI_TestSolutionServer) error {
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	util.Info("Received request to judge " + strconv.FormatInt(int64(req.Solution.Problem.ProblemID), 10) + " in " + req.Solution.Language + ".")

	stat := make(chan shared.JudgeStatus)
	streamReceive := make(chan pb.TestSolutionRequest)

	// whether or not the judge task has been cancelled (so that judging process exists)
	isCancelled := false
	defer func() { isCancelled = true }()

	// start judging
	go judging.TestSolution(req, stat, &isCancelled)

	// listen for stream input
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

	for { // listen for further requests and status simultaneously
		select {
		case s := <-stat: // if status update

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
