package goQA

import (
	"fmt"
	//"error"
	//"log"
	//	"os"
	"bufio"
	"encoding/json"
	"net"
	"time"
	//"../../goQA"
)

// send messages
const (
	CMD_UNKNOWN = iota
	CMD_OK
	CMD_NO_COMMAND
	CMD_LAUNCH_RUN
	CMD_CANCEL_CURRENT_RUN
	CMD_CANCEL_RUNPLAN
	CMD_NEW_BUILD
	CMD_GET_RUN_STATUS
	CMD_GET_RUNPLAN_STATUS
	CMD_GET_ALL_RUNPLAN_STATUS
	CMD_ERR_TIMEOUT
	CMD_ERR_UNKNOWN
)

const (
	LOOP_WAIT_TIMER    = 500
	COMMAND_QUEUE_SIZE = 100
)

const (
	MES_ADDRESS = "localhost"
	MES_PORT    = 1414
)

func CmdName(cmd int) string {
	var ret string
	switch cmd {
	case CMD_OK:
		ret = "OK"
	case CMD_NO_COMMAND:
		ret = "NO_COMMAND"
	case CMD_LAUNCH_RUN:
		ret = "LAUNCH_RUN"
	case CMD_CANCEL_CURRENT_RUN:
		ret = "CANCEL_CURRENT_RUN"
	case CMD_CANCEL_RUNPLAN:
		ret = "CANCEL_CURRENT_RUN"
	case CMD_NEW_BUILD:
		ret = "NEW_BUILD"
	case CMD_GET_RUN_STATUS:
		ret = "GET_RUN_STATUS"
	case CMD_GET_RUNPLAN_STATUS:
		ret = "GET_RUNPLAN_STATUS"
	case CMD_GET_ALL_RUNPLAN_STATUS:
		ret = "GET_ALL_RUNPLAN_STATUS"
	default:
		ret = "UNKNOWN"
	}
	return ret
}

type myError struct {
	mes string
}

func (e *myError) Error() string {
	return e.mes
}

type CommandInfo struct {
	Command int
	Data    []interface{}
}

type InternalCommandInfo struct {
	Command   int
	ChnReturn chan CommandInfo
	Data      []interface{}
}

type CommandInfoEncoder interface {
	Unmarshal(data []byte) (CommandInfo, error)
	Marshal(data interface{}) ([]byte, error)
}

type JSON_Encode struct {
	rsvMessage CommandInfo
}

func (j *JSON_Encode) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (j *JSON_Encode) Unmarshal(data []byte) (CommandInfo, error) {
	err := json.Unmarshal(data, &j.rsvMessage)
	return j.rsvMessage, err
}

type MessageSender struct {
	logger  *GoQALog
	conn    net.Conn
	Port    int
	Address string
	encoder CommandInfoEncoder
}

func (snd *MessageSender) Init(address string, port int, logger *GoQALog, encoder CommandInfoEncoder) {
	snd.logger = logger
	snd.Address = address
	snd.Port = port
	snd.encoder = encoder
}

func (snd *MessageSender) Connect() error {
	if snd.conn != nil {
		snd.conn.Close()
		snd.conn = nil
	}
	addr := fmt.Sprintf("%s:%d", snd.Address, snd.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		snd.logger.LogError("Error connecting to ....")
		return err
	}
	snd.conn = conn
	return nil
}

func (snd *MessageSender) isConnected() bool {
	if snd.conn != nil {
		return true
	} else {
		return false
	}
}

func (snd *MessageSender) Send(command int, data ...interface{}) (CommandInfo, error) {
	var status string
	var sendData []byte
	var rsvMessage CommandInfo
	var err error

	if err = snd.Connect(); err != nil {
		snd.logger.LogError("Send()::ERROR:: %s", err)
		return CommandInfo{}, err
	}
	commandInfo := GetMessageInfo(command, data...)
	sendData, err = snd.encoder.Marshal(commandInfo)
	if err != nil {
		snd.logger.LogError("Error: ", err)
	}
	snd.logger.LogDebug("CLIENT Send data::%s", sendData)
	fmt.Fprintf(snd.conn, string(sendData))
	snd.logger.LogDebug("CLIENT RESVing....")
	status, err = bufio.NewReader(snd.conn).ReadString('\n')
	if err != nil {
		snd.logger.LogError("Error: ", err)
	}
	rsvMessage, err = snd.encoder.Unmarshal([]byte(status))
	snd.logger.LogDebug("CLIENT recieved::%s", status)
	snd.conn.Close()
	snd.conn = nil
	return rsvMessage, nil
}

type MessageReciever struct {
}

func (rcv *MessageReciever) Connect(address string, port int) {

}

type CommandQueue chan InternalCommandInfo

type Master struct {
	logger             *GoQALog
	chnExit            chan int
	m_runplanRouter    RunplanRouter
	chnRunplanRouter chan RunInfo
	m_chnScheduler     chan InternalCommandInfo
	m_runQueue         chan RunInfo
	m_commandQueue     *CommandQueue
}

// convenience methods to call the logger methods

func (m *Master) SetDebug(mode bool) {
	m.logger.SetDebug(mode)
}

func (m *Master) LogError(errMsg string, args ...interface{}) {
	m.logger.LogError(errMsg, args...)
}

func (m *Master) LogDebug(debugMsg string, args ...interface{}) {
	m.logger.LogDebug(debugMsg, args...)
}

func (m *Master) LogWarning(warnMsg string, args ...interface{}) {
	m.logger.LogWarning(warnMsg, args...)
}

func (m *Master) LogMessage(msg string, args ...interface{}) {
	m.logger.LogMessage(msg, args...)
}

func (m *Master) Init(connector RemoteConnector, commandQueue *CommandQueue, chnExit chan int, logger *GoQALog) *Master {
	m.m_commandQueue = commandQueue
	m.m_runQueue = make(chan RunInfo, 100)
	m.chnExit = chnExit
	m.logger = logger
	m.chnRunplanRouter = make(chan RunInfo)
	m.StartMasterScheduler()
	go RunListener(connector, commandQueue, chnExit, logger)
	return m
}
func (m *Master) Stop(message int) bool {
	m.LogDebug("MASTER:STOP")
	return true
}

func (m *Master) Run() {
	isStopRequested := false

	for isStopRequested == false {
		select {
		case nextCommand := <-*m.m_commandQueue:
			go m.ProcessMessage(&nextCommand)
		case exitMessage := <-m.chnExit:
			isStopRequested = m.Stop(exitMessage)
		case <-time.After(time.Millisecond * LOOP_WAIT_TIMER):
		}
		m.onProcessEvents()
	}
	m.LogDebug("Out of master loop")
}

func (m *Master) GetCommandQueue() *CommandQueue {
	return m.m_commandQueue
}

func (m *Master) ProcessMessage(info *InternalCommandInfo) {
	m.LogDebug("CALLED->--ProcessMessage--")
	m.LogDebug("command: %s,  data: %s", CmdName(info.Command), info.Data)

	switch info.Command {
	case CMD_LAUNCH_RUN:
		m.LogMessage("launching runplan '%s'", info.Data[0])
		run := RunInfo{Id: info.Data[0].(RunId), Name: info.Data[1].(string), LaunchType: info.Data[2].(string)}
		m.chnRunplanRouter <- run
		mesRet := GetMessageInfo(CMD_OK, "Runplan launched", time.Now().String())
		info.ChnReturn <- mesRet
	default:
		m.LogMessage("MASTER::No Command -> %s %v %v", CmdName(info.Command), info.Data, &info.ChnReturn)
		mesRet := GetMessageInfo(CMD_NO_COMMAND, "No command processed")
		info.ChnReturn <- mesRet
	}
}

func (m *Master) onProcessEvents() {
	m.LogDebug("CALLED->onProcessEvents")
}

func (m *Master) StartMasterScheduler() error {
	m.LogDebug("MASTER:Creating scheduler....")
	m.m_runplanRouter = RunplanRouter{}
	m.m_runplanRouter.Init(m.chnRunplanRouter, m.chnExit, m.logger)

	chnScheduler := make(chan *RunInfo)
	sched := Schedule{}
	sched.Init(chnScheduler, m.logger, "Default Scheduler")
	sched.Start()
	schedHandler := SimpleSchedueHandler{}
	schedHandler.Init(chnScheduler)
	m.m_runplanRouter.AddScheduleHandler("Defult Schedule Handler", &schedHandler)
	m.m_runplanRouter.Run()
	return nil
}

func GetMessageInfo(cmd int, data ...interface{}) CommandInfo {
	return CommandInfo{Command: cmd, Data: data}
}

func GetInternalMessageInfo(cmd int, chnRet chan CommandInfo, data ...interface{}) InternalCommandInfo {
	return InternalCommandInfo{Command: cmd, ChnReturn: chnRet, Data: data}
}
