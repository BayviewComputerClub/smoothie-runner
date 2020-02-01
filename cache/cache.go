package cache

import (
	"encoding/base64"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/test-data"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Meta struct {
	Hash string `yaml:"hash"`
}

func InitCache() {
	src, err := os.Stat(shared.CACHE_DIR)
	if os.IsNotExist(err) {
		err := os.MkdirAll(shared.CACHE_DIR, 0755)
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
	confRaw, err := ioutil.ReadFile(folderName + "meta.yml")
	if err != nil {
		return false
	}
	err = yaml.Unmarshal([]byte(confRaw), &conf)
	if err != nil {
		return false
	}

	return conf.Hash == hash
}

// get from cache
func GetTestData(problemId string) pb.

func getBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}