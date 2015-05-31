package main

import (
	//"fmt"
	//"error"
	//"log"
	"os"
	"runtime"
	//"io"
	//"gitorious.org/goqa/goqa.git"
	"../../goQA"
	"time"
	//"net"
	//"encoding/json"
)

func main() {
	
 	console, err := os.Create("data/console.log")
	if err != nil { panic(err) }
	defer console.Close()

	errLog, err := os.Create("data/error.log")
	if err != nil { panic(err) }
	defer errLog.Close()

	incedentLog, err := os.Create("data/incedents.log")
	if err != nil { panic(err) }
	defer incedentLog.Close()

	resultLog, err := os.Create("data/TestResults.log")
	if err != nil { panic(err) }
	defer resultLog.Close()

	logger := goQA.GoQALog{}
	logger.Init()
	logger.SetDebug(true)

	logger.Add("default", goQA.LOGLEVEL_ALL, os.Stdout)
	logger.Add("Console", goQA.LOGLEVEL_MESSAGE, console)
	logger.Add("Error", goQA.LOGLEVEL_ERROR, errLog)
	logger.Add("Incidents", goQA.LOGLEVEL_WARNING, incedentLog)
	logger.Add("Resuts", goQA.LOGLEVEL_PASS_FAIL, resultLog)

	logger.LogMessage("running on platform %s", runtime.GOOS)
	logger.LogMessage("First message")
	logger.LogMessage("second message")
	logger.LogMessage("third message")
	logger.LogDebug("Debug message")
	logger.LogWarning("Warning Will Robinson")
	logger.LogPass("Test Passed")
	logger.LogFail("Test Failed")
	logger.LogError("Failure in script")

	time.Sleep(time.Second * 1)
}
