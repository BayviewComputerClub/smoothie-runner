package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	runner "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	testData "github.com/BayviewComputerClub/smoothie-runner/protocol/test-data"
	"github.com/golang/protobuf/proto"
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
	GRADER               *string

	PROBLEM_ID                *string
	TEST_BATCH_EVEN_IF_FAILED *bool
	UPLOAD_TEST_DATA          *string
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
	GRADER = flag.String("grader", "strict", "specify the grader to use")
	PROBLEM_ID = flag.String("problemid", "1", "specify the problem id")
	TEST_BATCH_EVEN_IF_FAILED = flag.Bool("forcetest", false, "test batch even if failed")
	UPLOAD_TEST_DATA = flag.String("upload", "true", "whether or not to update test data cache")

	flag.Parse()

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

	client := runner.NewSmoothieRunnerAPIClient(conn)

	// create test data
	test := &testData.TestData{
		Batch: []*testData.TestDataBatch{
			{
				Case: []*testData.TestDataBatchCase{
					{
						Input:          string(input),
						ExpectedOutput: string(output),
						BatchNum:       0,
						CaseNum:        0,
					},
				},
				BatchNum: 0,
			},
		},
	}

	testBytes, err := proto.Marshal(test)
	if err != nil {
		log.Fatal(err)
	}

	hash := md5.Sum(testBytes)
	testDataStream, err := client.UploadProblemTestData(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if *UPLOAD_TEST_DATA == "true" {
		log.Println("Connected! Sending test data...")

		// send test data first

		err = testDataStream.Send(&runner.UploadTestDataRequest{
			DataChunk:         testBytes,
			ProblemId:         *PROBLEM_ID,
			TestDataHash:      fmt.Sprintf("%x", hash),
			FinishedUploading: true,
		})
		if err != nil {
			log.Fatal(err)
		}

		msg, err := testDataStream.CloseAndRecv()
		if err != nil {
			log.Fatal(err)
		}
		if msg.Error != "" {
			log.Fatal("Test data upload error: " + msg.Error)
		}
	}

	log.Println("Finished! Sending solution...")

	// sending requests
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

			if in.TestDataNeedUpload {
				log.Println("Test data needs uploading.")
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

	err = stream.Send(&runner.TestSolutionRequest{
		Solution: &runner.Solution{
			Language: *LANGUAGE,
			Code:     string(code),
		},
		Problem: &runner.Problem{
			ProblemId:    *PROBLEM_ID,
			TestDataHash: fmt.Sprintf("%x", hash),
			Grader: &runner.ProblemGrader{
				Type:       *GRADER,
				CustomCode: "",
			},
			TimeLimit: *TIME_LIMIT,
			MemLimit:  *MEM_LIMIT,
		},
		TestBatchEvenIfFailed: *TEST_BATCH_EVEN_IF_FAILED,
		CancelTesting:         false,
	})

	log.Println("Request sent!")

	_ = stream.CloseSend()
	<-waitc
}
