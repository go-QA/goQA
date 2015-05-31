// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
	//"sync"
	//"error"
	"log"
	"runtime"
	//"os"
	"io"
	"time"
)

const (
	LOG_QUEUE_SIZE = 100
	LOG_SYNC_DELAY = 2
)

const LOGLEVEL_NOT_SET = 0
const (
	LOGLEVEL_DEBUG = (1 << iota)
	LOGLEVEL_MESSAGE
	LOGLEVEL_WARNING
	LOGLEVEL_PASS_FAIL
	LOGLEVEL_ERROR
	LOGLEVEL_ALL
)

const log_UNKNOWN = 0
const (
	log_DEBUG = (1 << iota)
	log_MESSAGE
	log_WARNING
	log_PASS
	log_FAIL
	log_ERROR
)

type logArg struct {
	level   uint64
	pattern string
	args    []interface{}
}

type logStream struct {
	ChnLogInput chan logArg
	level       uint64
	logger      log.Logger
}

func (log *logStream) Init(debug bool) {

	log.ChnLogInput = make(chan logArg, LOG_QUEUE_SIZE)

	go func(bool) {
		for message := range log.ChnLogInput {
			loggerLevel := log.level
			mesLevel := message.level
			switch {
			case loggerLevel == uint64(LOGLEVEL_ALL):
				log.logger.Printf(message.pattern, message.args...)
			case (loggerLevel == LOGLEVEL_MESSAGE):
				if (mesLevel != log_DEBUG) || ((mesLevel == log_DEBUG) && (debug == true)) {
					log.logger.Printf(message.pattern, message.args...)
				}
			case (loggerLevel == uint64(LOGLEVEL_WARNING)) && ((mesLevel & (log_WARNING | log_ERROR)) != 0):
				log.logger.Printf(message.pattern, message.args...)
			case (loggerLevel == LOGLEVEL_ERROR) && (mesLevel == log_ERROR):
				log.logger.Printf(message.pattern, message.args...)
			case (loggerLevel == uint64(LOGLEVEL_PASS_FAIL)) && ((mesLevel & (log_PASS | log_FAIL)) != 0):
				log.logger.Printf(message.pattern, message.args...)
			}
		}
	}(debug)

}

func (log *logStream) sync() {
	for len(log.ChnLogInput) > 0 {
		time.Sleep(time.Millisecond * LOG_SYNC_DELAY)
	}
}

type GoQALog struct {
	chnInput    chan logArg
	loggers     map[string]logStream
	debugMode   bool
	initialized bool
	faulted     bool
	END         string
}

// Init method will automatically be called before logger is used but user can call if desired.
func (gLog *GoQALog) Init() {
	if gLog.initialized {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			gLog.faulted = true
			panic(r)
		}
	}()
	gLog.loggers = make(map[string]logStream)
	gLog.chnInput = make(chan logArg, LOG_QUEUE_SIZE)
	gLog.debugMode = false
	if runtime.GOOS == "linux" {
		gLog.END = "\n"
	} else {
		gLog.END = "\r\n"
	}
	gLog.initialized = true

	go func() {
		for message := range gLog.chnInput {
			for _, logger := range gLog.loggers {
				logger.ChnLogInput <- message
			}
		}
	}()
}

func (gLog *GoQALog) ready() bool {
	if gLog.initialized {
		return true
	}
	if gLog.faulted {
		return false
	}
	gLog.Init()
	return gLog.initialized
}

func (gLog *GoQALog) Add(name string, level uint64, stream io.Writer) {
	if !gLog.ready() {
		return
	}
	if _, ok := gLog.loggers[name]; !ok {
		stream := logStream{level: level, logger: *log.New(stream, "", log.Ldate|log.Ltime|log.Lmicroseconds)}
		stream.Init(gLog.debugMode)
		gLog.loggers[name] = stream
	}
}

func (gLog *GoQALog) Printf(level uint64, value string, args ...interface{}) {
	arg := logArg{level, value, args}
	gLog.chnInput <- arg
}

func (gLog *GoQALog) Sync() {
	if !gLog.ready() {
		return
	}
	for len(gLog.chnInput) > 0 {
		time.Sleep(time.Millisecond * LOG_SYNC_DELAY)
	}
	for _, logger := range gLog.loggers {
		logger.sync()
	}
}
func (gLog *GoQALog) SetDebug(mode bool) {
	if gLog.ready() {
		gLog.debugMode = mode
	}
}

func (gLog *GoQALog) LogError(errMsg string, args ...interface{}) {
	if gLog.ready() {
		gLog.Printf(log_ERROR, fmt.Sprintf("ERROR::%s%s", errMsg, gLog.END), args...)
	}
}

func (gLog *GoQALog) LogDebug(DebugMsg string, args ...interface{}) {
	if gLog.ready() {
		if gLog.debugMode {
			gLog.Printf(log_DEBUG, fmt.Sprintf("DEBUG::%s%s", DebugMsg, gLog.END), args...)
		}
	}
}

func (gLog *GoQALog) LogWarning(warnMsg string, args ...interface{}) {
	if gLog.ready() {
		gLog.Printf(log_WARNING, fmt.Sprintf("ERROR::%s%s", warnMsg, gLog.END), args...)
	}
}

func (gLog *GoQALog) LogPass(passMsg string, args ...interface{}) {
	if gLog.ready() {
		gLog.Printf(log_PASS, fmt.Sprintf("PASS::%s%s", passMsg, gLog.END), args...)
	}
}

func (gLog *GoQALog) LogFail(failMsg string, args ...interface{}) {
	if gLog.ready() {
		gLog.Printf(log_FAIL, fmt.Sprintf("FAIL::%s%s", failMsg, gLog.END), args...)
	}
}

func (gLog *GoQALog) LogMessage(msg string, args ...interface{}) {
	if gLog.ready() {
		gLog.Printf(log_MESSAGE, fmt.Sprintf("MSG::%s%s", msg, gLog.END), args...)
	}
}
