package main

import (
	"fmt"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
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
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}


	}
}

// rpc server start

func startApiServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		fatal("IPC listen error (check if port has been taken):" + err.Error())
	}
	pb.RegisterSmoothieRunnerAPIServer(smoothieRunner, &SmoothieRunnerAPI{})
	info("Started API on port " + strconv.Itoa(PORT))
	go smoothieRunner.Serve(lis)
}