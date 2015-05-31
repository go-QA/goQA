package goQA

import (
	"fmt"
	//"error"
	//"log"
	//"os"
	//"encoding/json"
	"time"
	"net"
	//"bufio"
	//"../../goQA"
)

//Listener interface used as a connector to external information.
// The Start() is used to return a channel to pass information to 
// am oject. The Listener recieves a return response 
type Listener interface {

	Init(logger *GoQALog, encoder CommandInfoEncoder, args ...interface{})
	//The Start needs to return a pointer to an InternalCommandInfo object.
	// Send data to the channel
	// recieve the ddata that will be returned on the recieve channel of InternalCommandInfo
	Start(mesChan chan<- *InternalCommandInfo)  

	//The Listener Will stop sending messages and not recieve anymore
	Stop()
}

// Remote  connector takes a Listener object to get information and can send it to 
// the CommandQueue passed in.
type RemoteConnector interface {
	Init(listen Listener, commandQueue *CommandQueue, chnExit chan int, logger *GoQALog)
	// Run is expected to call the Listener objects Run() that returns a pointer to
	// InternalCommandInfo. This is used to recieve the commands and return 
	// messages through the return channel back to Listener
	Run()
	// Stop sending and recieving messages when called. The object should exit its
	// main loop and close all channels.
	Stop()
}

// TCPConnector is a concrete Listener that uses the goQA message protocal
// to send and recieve messages over TCP/IP
type TCPConnector struct {
	logger *GoQALog
	err error
	conn net.Conn
	m_listener net.Listener
	encoder CommandInfoEncoder
	port int
	address string
	buffer []byte
	FullAddress string
}
// Init takes 3 arguments to initialize socket: 
//    adress string
//    port int
//    encoder CommandInfoEncoder
func (mes *TCPConnector) Init(logger *GoQALog, encoder CommandInfoEncoder, args ...interface{} ) {
	mes.logger = logger
	mes.encoder = encoder
	mes.address = args[0].(string)
	mes.port = args[1].(int)
	mes.buffer = make([]byte, 512)
	mes.FullAddress = fmt.Sprintf("%s:%d", mes.address, mes.port)
	mes.logger.LogMessage("Address for listening:%s\n", mes.FullAddress)
	mes.m_listener, mes.err = net.Listen("tcp", mes.FullAddress)
	if mes.err != nil {
		mes.logger.LogError("Failed to create listener socket")
		panic(mes.err)
	}

}

func (mes *TCPConnector) Start(mesChan chan<- *InternalCommandInfo) {
	//mesChn := make (chan *InternalCommandInfo)
	var nextMessage *InternalCommandInfo
	var conn net.Conn
	//retChn := make(chan CommandInfo)
	go func() {
		for {
			mes.logger.LogDebug("TCPConnector::Getting message::")
			nextMessage, conn = mes.getNextMessage()
			mes.logger.LogDebug("TCPConnector::Mes got::%s", nextMessage.Command)
			mesChan <- nextMessage
			//retMessage := <-nextMessage.ChnReturn
			go mes.returnMessage(nextMessage, conn)

		}	
	}()
}

func (mes *TCPConnector)  getNextMessage() (*InternalCommandInfo, net.Conn) {
	var mesRecieved InternalCommandInfo
	var inMessage CommandInfo
	conn, err := mes.m_listener.Accept()
	mes.logger.LogDebug("new connection recieved")
	if err != nil {
		mes.logger.LogError("RUN::conn except error::", err)
		panic(err)
	}
	bytelength, readErr := conn.Read(mes.buffer)
	if readErr != nil {
		mes.logger.LogError("RUN::error unmarsheling com info")
		panic(readErr)
	}
	mes.logger.LogDebug("mess Rec::")
	//fmt.Println("byteLength=", bytelength)
	//trimmedString := string(buffer[0:bytelength])
	//fmt.Printf("MSG:%s\n", trimmedString)
	inMessage, mes.err = mes.encoder.Unmarshal(mes.buffer[0:bytelength])
	mesRecieved = GetInternalMessageInfo(inMessage.Command, make(chan CommandInfo), inMessage.Data...)
	if mes.err != nil {
		mes.logger.LogError("RUN::error unmarsheling com info")
		panic(mes.err)
	}
	return &mesRecieved, conn
}

func (mes *TCPConnector)  returnMessage(chRetMessage *InternalCommandInfo, conn net.Conn) {
	var retInfo []byte
	var mesSend CommandInfo
	select {
		case retMessage := <-chRetMessage.ChnReturn:
			mesSend = GetMessageInfo(retMessage.Command, retMessage.Data...)
		case <-time.After(time.Minute * 10):
			mes.logger.LogError("Took to long to return message")
			mesSend = GetMessageInfo(CMD_ERR_TIMEOUT, "Took to long to return message")
	}
	retInfo, mes.err = mes.encoder.Marshal(mesSend)
	mesReturn := string(retInfo)
	mesReturn = mesReturn + "\n"
	fmt.Fprintf(conn, mesReturn)
	conn.Close()
	conn = nil
}

func (mes *TCPConnector) Stop() {

}


//InternalConnector is a concrete object that takes a listener. It will read each
// message from InternalCommandInfo provided by listener and send the message to the 
// CommandQueue. It then waits for the return massage and send back to listener.
type ExternalConnector struct {
	logger *GoQALog
	m_commandQueue *CommandQueue
	m_listener Listener
	m_chnListener chan *InternalCommandInfo
}

func (l *ExternalConnector) Init(listener Listener, commandQueue *CommandQueue, chnExit chan int, logger *GoQALog) {
	l.logger = logger
	l.m_commandQueue = commandQueue
	l.m_listener = listener
}

func (l *ExternalConnector) Stop() {
	
}

func (l *ExternalConnector) Run() {
	var err error
	var mesRecieved *InternalCommandInfo
	var mesToSend    InternalCommandInfo
	var isMessageRecieved bool
	
	mesChn := make (chan *InternalCommandInfo)
	l.m_listener.Start(mesChn)

	for {
		isMessageRecieved = false
		for isMessageRecieved == false {
			select {
				case mesRecieved = <- mesChn:
					isMessageRecieved = true
					mesToSend = GetInternalMessageInfo(mesRecieved.Command, make(chan CommandInfo), mesRecieved.Data...)
					*l.m_commandQueue <-mesToSend
					l.logger.LogDebug("Wait return...%s", &mesToSend.ChnReturn)
				case  <- time.After(time.Second * 10):
					l.logger.LogDebug("No message recieved for long time")
			}
		}
		if isMessageRecieved == true {
			select {
				case returned := <- mesToSend.ChnReturn:
					l.logger.LogDebug("Listener resved %s %s", CmdName(returned.Command), returned.Data )
					//err = messageListener.ReturnMessage(returned.Command, returned.Data...)
					mesRecieved.ChnReturn <-GetMessageInfo(returned.Command, returned.Data...)
					if err != nil {
						l.logger.LogDebug("ERROR::%s", err)
					}		
				case  <- time.After(time.Second * 10):
					l.logger.LogDebug("Return message took too long")
					//err = messageListener.ReturnMessage(CMD_NO_COMMAND, "TImed out")
					mesRecieved.ChnReturn <-GetMessageInfo(CMD_NO_COMMAND, "TImed out")		
					if err != nil {
						l.logger.LogDebug("ERROR::", err)
					}		
			}
		}
	}
}


type Reciever interface {
	Start() (*chan InternalCommandInfo, error)
}

type Recieve struct {
	m_chnMessage chan InternalCommandInfo
}

func (r *Recieve ) Start() (chan InternalCommandInfo, error) {
	r.m_chnMessage = make(chan InternalCommandInfo)
	return r.m_chnMessage, nil
}


func RunListener(connector RemoteConnector, commandQueue *CommandQueue, chnExit chan int,  logger *GoQALog) {
	m_chnExit := chnExit
	//connector.Init(listener, commandQueue, chnExit, logger)
	go connector.Run()
	_ = <-m_chnExit
	connector.Stop()
	logger.LogMessage("Leaving RunListener")

}

