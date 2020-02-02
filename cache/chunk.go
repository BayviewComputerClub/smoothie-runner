package cache

import (
	"errors"
	testdata "github.com/BayviewComputerClub/smoothie-runner/protocol/test-data"
	"github.com/golang/protobuf/proto"
)

var (
	DataChunks = make(map[string][]byte) // temporarily store test data that is being streamed from client
)

// add byte chunk of a pb.TestData object
// NOT THREAD SAFE
func AddByteChunk(problemId string, chunk []byte) {
	if v, ok := DataChunks[problemId]; ok {
		DataChunks[problemId] = append(v, chunk...)
	} else {
		DataChunks[problemId] = chunk
	}
}

func AddToCacheFromChunks(problemId string, hash string) error {
	if _, ok := DataChunks[problemId]; !ok {
		return errors.New("problem chunks not found")
	}

	testData := testdata.TestData{}
	err := proto.Unmarshal(DataChunks[problemId], &testData)
	if err != nil {
		return err
	}

	return AddToCache(problemId, hash, testData)
}
