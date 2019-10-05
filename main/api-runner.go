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

	stat := make(chan shared.JudgeStatus)

	// whether or not the judge task has been cancelled (so that judging process exists)
	isCancelled := false
	defer func() { isCancelled = true }()

	// start judging
	go judging.TestSolution(req, stat, &isCancelled)

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

		default: // if no status update, read from stream
			d, err := stream.Recv()
			if err == io.EOF {
				return nil // TODO
			}
			if err != nil {
				return err
			}

			if d.CancelTesting {
				return nil
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
