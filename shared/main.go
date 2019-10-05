package shared

import (
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol"
)

var (
	PORT int
	MAX_CORES int
)

const (
	OUTCOME_AC  = "AC"  // answer correct
	OUTCOME_WA  = "WA"  // wrong answer
	OUTCOME_RTE = "RTE" // run time error
	OUTCOME_CE  = "CE"  // compile error
	OUTCOME_TLE = "TLE" // time limit exceeded
	OUTCOME_MLE = "MLE" // memory limit exceeded
	OUTCOME_ILL = "ILL" // illegal operation
	OUTCOME_ISE = "ISE" // internal server error
)

type JudgeStatus struct {
	Err error // any possible errors
	Res pb.TestSolutionResponse // response to send back
}