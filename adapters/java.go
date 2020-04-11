package adapters

import (
	"errors"
	pb "github.com/BayviewComputerClub/smoothie-runner/protocol/runner"
	"github.com/BayviewComputerClub/smoothie-runner/shared"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// taken from https://github.com/DMOJ/judge-server/blob/master/dmoj/executors/java-security.policy
var JavaSecurityPolicy = `
grant {
    // Read/write to system streams
    permission java.lang.RuntimePermission "readFileDescriptor";
    permission java.lang.RuntimePermission "writeFileDescriptor";

    // Basic Threads
    permission java.lang.RuntimePermission "stopThread";
    permission java.lang.RuntimePermission "modifyThread";

    // Locale.setDefault
    permission java.util.PropertyPermission "user.language", "write";

    // Date timezone
    permission java.util.PropertyPermission "user.timezone", "write";

    // Standard properties
    permission java.util.PropertyPermission "java.version", "read";
    permission java.util.PropertyPermission "java.vendor", "read";
    permission java.util.PropertyPermission "java.vendor.url", "read";
    permission java.util.PropertyPermission "java.class.version", "read";
    permission java.util.PropertyPermission "os.name", "read";
    permission java.util.PropertyPermission "os.version", "read";
    permission java.util.PropertyPermission "os.arch", "read";
    permission java.util.PropertyPermission "line.separator", "read";

    permission java.util.PropertyPermission "java.specification.version", "read";
    permission java.util.PropertyPermission "java.specification.vendor", "read";
    permission java.util.PropertyPermission "java.specification.name", "read";

    permission java.util.PropertyPermission "java.vm.specification.version", "read";
    permission java.util.PropertyPermission "java.vm.specification.vendor", "read";
    permission java.util.PropertyPermission "java.vm.specification.name", "read";
    permission java.util.PropertyPermission "java.vm.version", "read";
    permission java.util.PropertyPermission "java.vm.vendor", "read";
    permission java.util.PropertyPermission "java.vm.name", "read";
};
`

type Java11Adapter struct{}

func (adapter Java11Adapter) GetName() string {
	return "java11"
}

func (adapter Java11Adapter) Compile(session *shared.JudgeSession) (*exec.Cmd, error) {
	return JavaHelper(session)
}

func (adapter Java11Adapter) JudgeFinished(tcr *pb.TestCaseResult) {
	if tcr.Result == shared.OUTCOME_RTE {
		// mle
		if strings.Contains(tcr.ResultInfo, "There is insufficient memory for the Java Runtime Environment") ||
			strings.Contains(tcr.ResultInfo, "java.lang.OutOfMemoryError: Java heap space"){
			tcr.Result = shared.OUTCOME_MLE
			//tcr.ResultInfo = ""
		}
	}
}

func JavaHelper(session *shared.JudgeSession) (*exec.Cmd, error) {
	//session.FSizeLimit = 64 // dump file
	session.NProcLimit = -1 // infinite threads (jvm)
	// set memory limit to zero since it's enforced by the jvm
	session.MemLimit = 0

	// write source file
	err := ioutil.WriteFile(session.Workspace+"/Main.java", []byte(session.Code), 0644)
	if err != nil {
		return nil, err
	}

	// write security policy
	err = ioutil.WriteFile(session.Workspace+"/policy", []byte(JavaSecurityPolicy), 0644)
	if err != nil {
		return nil, err
	}

	// compile
	output, err := exec.Command("javac", session.Workspace+"/Main.java").CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.ReplaceAll(string(output), session.Workspace+"/Main.java", ""))
	}

	// get current working directory
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// java agent option for sandboxing
	javaagent := "-javaagent:"+path+"/"+shared.JAVA_SANDBOX_AGENT+"=policy:policy"
	if !shared.SANDBOX {
		javaagent = ""
	}

	// command for execution
	c := exec.Command("java",
		javaagent,
		"-Xmx"+strconv.Itoa(int(session.OriginalRequest.Problem.MemLimit))+"M",
		"-Xss128m", "-XX:+UseSerialGC", "-XX:ErrorFile=crash.log",
		"Main")

	c.Env = append(c.Env, "MALLOC_ARENAS_MAX=1")
	c.Dir = session.Workspace

	// no seccomp
	session.SandboxWithSeccomp = false

	return c, nil
}
