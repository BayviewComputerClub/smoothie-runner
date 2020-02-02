package cache

import (
	"encoding/base64"
	"fmt"
	test_data "github.com/BayviewComputerClub/smoothie-runner/protocol/test-data"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Meta struct {
	Hash string `yaml:"hash"`
}

type CachedTestDataCase struct {
	Input    *os.File
	Output   *os.File // expected output
	CaseNum  int
	BatchNum int
}

type CachedTestDataBatch struct {
	Cases    []CachedTestDataCase
	BatchNum int
}

type CachedTestData struct {
	Batches []CachedTestDataBatch
}

func InitCache() {
	src, err := os.Stat(shared.CACHE_DIR)
	if os.IsNotExist(err) {
		err := os.MkdirAll(shared.CACHE_DIR, 0644)
		if err != nil {
			panic(err)
		}
	}
	if src.Mode().IsRegular() {
		log.Fatal("Tried to create cache folder, but it exists as a file!")
	}
}

// check if the problem exists, and the hash matches
func Match(problemId string, hash string) bool {
	folderName := shared.CACHE_DIR + "/" + getBase64(problemId)

	// check if problem folder exists
	_, err := os.Stat(folderName)
	if os.IsNotExist(err) {
		return false
	}

	// get hash
	var conf Meta
	confRaw, err := ioutil.ReadFile(folderName + "/meta.yml")
	if err != nil {
		return false
	}
	err = yaml.Unmarshal(confRaw, &conf)
	if err != nil {
		return false
	}

	return conf.Hash == hash
}

func AddToCache(problemId string, hash string, testData test_data.TestData) error {
	folderName := shared.CACHE_DIR + "/" + getBase64(problemId)

	_, err := os.Stat(folderName)
	if !os.IsNotExist(err) {
		os.RemoveAll(folderName)
	}

	err = os.MkdirAll(folderName, 0644)
	if err != nil {
		return err
	}

	// make meta file with hash and add to folder
	conf := Meta{Hash: hash}
	str, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(folderName + "/meta.yml", str, 0644)
	if err != nil {
		return err
	}

	// add test data
	for _, b := range testData.Batch {
		for _, c := range b.Case {
			err = ioutil.WriteFile(fmt.Sprintf("%s/%d-%d.in", folderName, c.BatchNum, c.CaseNum), []byte(c.Input), 0644)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(fmt.Sprintf("%s/%d-%d.out", folderName, c.BatchNum, c.CaseNum), []byte(c.ExpectedOutput), 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// get from cache
func GetTestData(problemId string) (*CachedTestData, error) {
	folderName := shared.CACHE_DIR + "/" + getBase64(problemId)

	testData := CachedTestData{
		Batches: []CachedTestDataBatch{},
	}

	// loop files in problem cache directory
	err := filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		spl := strings.Split(info.Name(), ".")
		fType := spl[len(spl)-1]
		var f *os.File

		if fType == "in" { // input file
			f, err = os.Open(path)
			if err != nil {
				return err
			}
		} else if fType == "out" { // output file
			f, err = os.OpenFile(path, os.O_RDWR, os.ModeAppend) // open out file with read and write fd
			if err != nil {
				return err
			}
		} else { // a nothing file
			return nil
		}

		// format 0-0.in, 0-0.out, etc.
		spl2 := strings.Split(spl[0], "-")
		batchNum, err := strconv.Atoi(spl2[0])
		if err != nil {
			return err
		}
		caseNum, err := strconv.Atoi(spl2[1])
		if err != nil {
			return err
		}

		// add batches until batch num
		for len(testData.Batches) < batchNum+1 {
			testData.Batches = append(testData.Batches, CachedTestDataBatch{
				Cases:    []CachedTestDataCase{},
				BatchNum: len(testData.Batches),
			})
		}

		// add cases until case num
		for len(testData.Batches[batchNum].Cases) < caseNum+1 {
			testData.Batches[batchNum].Cases = append(testData.Batches[batchNum].Cases, CachedTestDataCase{
				Input:    nil,
				Output:   nil,
				CaseNum:  len(testData.Batches[batchNum].Cases),
				BatchNum: batchNum,
			})
		}

		// add case data
		if fType == "in" {
			testData.Batches[batchNum].Cases[caseNum].Input = f
		} else if fType == "out" {
			testData.Batches[batchNum].Cases[caseNum].Output = f
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &testData, nil
}

func getBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
