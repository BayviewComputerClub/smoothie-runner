package main

import (
	"context"
	"flag"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
)

var (
	PORT                 *string
	ADDRESS              *string
	LANGUAGE             *string
	CODE_FILE            *string
	INPUT_FILE           *string
	EXPECTED_ANSWER_FILE *string
	TIME_LIMIT           *float64
	MEM_LIMIT            *float64

	PROBLEM_ID                *uint64
	TEST_BATCH_EVEN_IF_FAILED *bool
)

func main() {
	// init flags
	PORT = flag.String("p", "6821", "specify the port of host")
	ADDRESS = flag.String("ip", "127.0.0.1", "specify the address of host")
	LANGUAGE = flag.String("language", "c++11", "specify the language to judge in")
	CODE_FILE = flag.String("codefile", "./main.cpp", "specify the location of the file that contains the code")
	INPUT_FILE = flag.String("inputfile", "./1.in", "specify the location of the file contains the input data")
	EXPECTED_ANSWER_FILE = flag.String("answerfile", "./1.out", "specify the location of the file that contains the expected output")
	TIME_LIMIT = flag.Float64("timelimit", 10.0, "specify the TLE time")
	MEM_LIMIT = flag.Float64("memlimit", 512, "specify the memory limit")
	PROBLEM_ID = flag.Uint64("problemid", 1, "specify the problem id")
	TEST_BATCH_EVEN_IF_FAILED = flag.Bool("forcetest", false, "test batch even if failed")

	// read input, output and code from file
	input, err := ioutil.ReadFile(*INPUT_FILE)
	if err != nil {
		log.Fatal(err.Error())
	}
	output, err := ioutil.ReadFile(*EXPECTED_ANSWER_FILE)
	if err != nil {
		log.Fatal(err.Error())
	}
	code, err := ioutil.ReadFile(*CODE_FILE)
	if err != nil {
		log.Fatal(err.Error())
	}

	// start connection
	log.Println("Attempting connection to host server...")
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*ADDRESS+":"+*PORT, opts...)

	if err != nil {
		log.Fatalf("Error connecting to host %s:%s, double check again?", *ADDRESS, *PORT)
	}

	client := pb.NewSmoothieRunnerAPIClient(conn)

	log.Println("Connected! Sending request...")

	stream, err := client.TestSolution(context.Background())

	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed receiving stream: %v", err)
			}

			log.Println("----- Request In -----")

			if in.CompletedTesting {
				log.Println("Completed testing.")
				log.Printf("Compile Error: %v\n", in.CompileError)
				log.Printf("Result: %v\n", in.TestCaseResult.Result)
				log.Println("-----------------------")
				close(waitc)
				return
			}

			log.Printf("Compile Error: %v\n", in.CompileError)
			log.Printf("Result: %v\n", in.TestCaseResult.Result)
			log.Printf("Result Info: %v\n", in.TestCaseResult.ResultInfo)
			log.Printf("Memusage: %v\n", in.TestCaseResult.MemUsage)
			log.Printf("Time: %v\n", in.TestCaseResult.Time)

			log.Println("-----------------------")
		}
	}()

	err = stream.Send(&pb.TestSolutionRequest{
		Solution: &pb.Solution{
			Problem: &pb.Problem{
				TestBatches: []*pb.ProblemBatch{{
					Cases: []*pb.ProblemBatchCase{{
						Input:          string(input),
						ExpectedAnswer: string(output),
						TimeLimit:      *TIME_LIMIT,
						MemLimit:       *MEM_LIMIT,
					}},
				}},
				ProblemID:         *PROBLEM_ID,
				TestCasesHashCode: 0,
			},
			Language: *LANGUAGE,
			Code:     string(code),
		},
		TestBatchEvenIfFailed: *TEST_BATCH_EVEN_IF_FAILED,
		CancelTesting:         false,
	})

	log.Println("Request sent!")

	_ = stream.CloseSend()
	<-waitc
}
