package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	"github.com/go-QA/logger"
	"time"
	//"net"
	//"encoding/json"
)

const (
	CLENT_DELAY1 = 1000
	CLENT_DELAY2 = 1000	
	BUILDER_COUNT = 3
	MESSAGE_COUNT = 2
)

const (
	MES_ADDRESS = "localhost"
	MES_PORT = 1414
)

func GetStatusRun(logger *goQA.GoQALog) {
	sender := goQA.MessageSender{}
	encoder := goQA.JSON_Encode{}
	sender.Init("localhost", 1414, logger, &encoder)
	for {
		time.Sleep(time.Millisecond * CLENT_DELAY1)
		rsvMessage, err := sender.Send(goQA.CMD_GET_RUN_STATUS, "ALL")
		if err == nil {
			logger.LogDebug("CLIENT RCV: %v\n", rsvMessage)
		} else {
			logger.LogDebug("CLIENT RCV ERROR::: ", err)
		}
		
	}
}

func BuildRun(chnBuild chan goQA.InternalCommandInfo, logger *goQA.GoQALog) {
	var buildInfo goQA.InternalCommandInfo
	verNum := 1
	for {		
		time.Sleep(time.Millisecond * CLENT_DELAY1)
		buildInfo = goQA.GetInternalMessageInfo(goQA.CMD_NEW_BUILD, make(chan goQA.CommandInfo), fmt.Sprintf("T1.0_%d", verNum), "Fun", "~/projects/fun", time.Now().String())
		verNum++
		chnBuild <- buildInfo
		go func(chnRet chan goQA.CommandInfo) {
			ret := <- chnRet
			logger.LogDebug("ClientRun::%s %s", goQA.CmdName(ret.Command), ret.Data[0])
			}(buildInfo.ChnReturn)
	}
}


func main() {

	chnExit := make(chan int)

	commandQueue := make(goQA.CommandQueue, 100)
	logger := goQA.GoQALog{}
	logger.Init()
	logger.Add("default", goQA.LOGLEVEL_ALL, os.Stdout)
	logger.SetDebug(true)

	messageListener := goQA.TCPConnector{}
	messageListener.Init(&logger, &goQA.JSON_Encode{}, MES_ADDRESS, MES_PORT)
	listener := goQA.ExternalConnector{}
	listener.Init(&messageListener, &commandQueue, chnExit, &logger)

	MockMatch := goQA.BuildMockMatch{}
	chnBuildIn := make(chan goQA.InternalCommandInfo)
	BuildMatcher := goQA.InternalBuildMatcher{}
	BuildMatcher.Init(&MockMatch, &chnBuildIn, &commandQueue, chnExit, &logger)

	master := goQA.Master{}
	master.Init(&listener, &commandQueue, chnExit, &logger)
	//time.Sleep(time.Second * 1)

	//go goQA.RunListener(&listener, &commandQueue, &chnExit)

	go BuildMatcher.Run()
	//time.Sleep(time.Second * 1)
	for i := 0; i < BUILDER_COUNT; i++ {
		go BuildRun(chnBuildIn, &logger)
		time.Sleep(time.Millisecond * 200)
	}
	for i := 0; i < MESSAGE_COUNT; i++ {
		go GetStatusRun(&logger)
		time.Sleep(time.Millisecond * 100)
	}

	logger.LogMessage("Running master...")
	master.Run()

	logger.LogMessage("Leaving Program....")
	time.Sleep(time.Second * 1)
}
