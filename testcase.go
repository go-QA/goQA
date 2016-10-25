// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
	"runtime"
	//"error"
	//"os"
	//"io"
	//"time"
	"github.com/go-QA/logger"
)

// Erro codes
const (
	_ = iota
	ErNone
	ErNotInitialized
	ErWrongData
)

// Results returned codes
const (
	_ = iota
	ResultPass
	ResultFail
	ResultWarning
	ResultError
)

// Concurrency levels for tests and suites
const (
	SuiteAll    = -1
	SuiteSerial = 0
	TcAll       = -1
	TcSerial    = 0
)

// TestError is object returned from test errors with extra
// information about test status and parameters used.
type TestError struct {
	stack   [4096]byte
	message string
	params  *Parameters
}

func (err *TestError) Error() string {
	return err.message
}

func (err *TestError) Create(params *Parameters, mes string) {
	err.params = params
	//err.message = mes
	runtime.Stack(err.stack[:], false)
	err.message = fmt.Sprintf("ERROR::%s\n", mes)
	err.message = fmt.Sprint("%sSTACK::%s\n", err.message, err.stack[:])
	if err.params.Count() > 0 {
		mes = fmt.Sprintf("%s\nPARMETERS::\n", err.message)
		for name, param := range err.params.params {
			mes = fmt.Sprintf("\t%s name: %s, value: %v, comment: %s\n", mes, name, param.value, param.comment)
		}
	}
	err.message = mes
}

func Create(params *Parameters, mes string) error {
	err := TestError{}
	err.Create(params, mes)
	return &err
}

func Throw(params *Parameters, mes string) {
	err := TestError{}
	err.Create(params, mes)
	panic(err)
}

type Section struct {
	active bool
	hits   int
}

func (s *Section) Trigger() {
	if s.active {
		s.hits++
	}
}

func (s *Section) Start() {
	s.active = true
}

func (s *Section) End() {
	s.active = false
}

func (s *Section) Triggered() bool {
	if s.hits > 0 {
		return true
	}
	return false
}

type Parameter struct {
	name    string
	value   interface{}
	comment string
}

func (p *Parameter) String() string {
	return fmt.Sprintf("%v", p.value)
}

type Parameters struct {
	params map[string]Parameter
}

func (p *Parameters) init() {
	if p.params == nil {
		p.params = make(map[string]Parameter)
	}
}

func (p *Parameters) Count() int {
	if p.params == nil {
		return 0
	}
	return len(p.params)
}

func (p *Parameters) updateValue(name string, value interface{}) interface{} {
	p.init()
	if _, ok := p.params[name]; ok {
		p.params[name] = Parameter{name, value, p.params[name].comment}
	}
	return p.params[name].value
}

func (p *Parameters) InitParam(name string, value interface{}) interface{} {
	p.init()
	if _, present := p.params[name]; !present {
		p.params[name] = Parameter{name, value, ""}
	} else {
		if p.params[name].value == nil && value != nil {
			p.updateValue(name, value)
		}
	}
	return p.params[name].value
}

// Add Param to the list of parameters. Will overwrite if already exists
// returns the value that is added as interface{}
func (p *Parameters) AddParam(name string, value interface{}, comment string) interface{} {
	p.init()
	p.params[name] = Parameter{name, value, comment}
	return p.params[name].value
}

// GetParam will return a Param object based on the param name passed in
// ok is True if the param exists. False if the param doesn't exist
func (p *Parameters) GetParam(name string) (Parameter, bool) {
	p.init()
	val, ok := p.params[name]
	return val, ok
}

func (p *Parameters) GetParamValue(name string) (interface{}, bool) {
	p.init()
	if param, ok := p.params[name]; !ok {
		return nil, false
	} else {
		return param.value, true
	}
}

func (p *Parameters) GetParamComment(name string) (interface{}, bool) {
	p.init()
	if param, ok := p.params[name]; !ok {
		return nil, false
	} else {
		return param.comment, true
	}
}

// the CreateParameters returns new empty Parameters object
func CreateParametersObject() Parameters {
	return Parameters{}
}

type iTestCase interface {
	Name() string
	Init(name string, parent ITestManager, params Parameters) ITestCase
	Setup() (int, error)
	Run() (int, error)
	Teardown() (int, error)
}

type ITestCase iTestCase
type ITestManager iTestManager

type TestCase struct {
	name   string
	parent iTestManager
	log    *logger.GoQALog
	//logChannel chan []byte
	params                                 Parameters
	failureThreshold                       int // percentage of check points that can fail for test case to passed
	passedCount, failedCount, warningCount int
	Critical                               Section
	//result TODO Define itss
	startTime float64
	endTime   float64
}

func (tc *TestCase) Name() string {
	return tc.name
}

func (tc *TestCase) Init(name string, parent ITestManager, params Parameters) ITestCase {
	tc.name = name
	tc.parent = parent
	tc.log = parent.GetLogger()
	tc.params = params
	tc.failureThreshold = tc.InitParam("failureThreshold", 0).(int)
	tc.Critical = Section{}
	return tc
}

func (tc *TestCase) Setup() (int, error) {
	return TcPassed, nil
}

func (tc *TestCase) Run() (int, error) {
	return tc.ReturnFromRun()
}

func (tc *TestCase) Teardown() (int, error) {
	return TcPassed, nil
}

func (tc *TestCase) Test(value int, mes string) (int, error) {
	return 0, nil
}

func (tc *TestCase) Check(value int, comment string, args ...interface{}) (int, error) {
	if value > 0 {
		tc.LogPass(fmt.Sprintf("CHECK::PASS::%s", comment), args...)
	} else {
		tc.LogFail(fmt.Sprintf("CHECK::FAIL::%s", comment), args...)
	}
	return value, nil
}

func (tc *TestCase) Verify(value bool, comment string, errMsg string, args ...interface{}) (bool, error) {
	if value {
		tc.LogPass(comment)
	} else {
		tc.LogFail(errMsg, args...)
	}
	return value, nil
}

func (tc *TestCase) GetLogger() *logger.GoQALog {
	return tc.log // tc.logChannel
}

func (tc *TestCase) LogError(errMsg string, args ...interface{}) {
	tc.Critical.Trigger()
	tc.failedCount++
	tc.log.LogError(errMsg, args...)
}

func (tc *TestCase) LogFail(failMsg string, args ...interface{}) {
	tc.Critical.Trigger()
	tc.failedCount++
	tc.log.LogFail(failMsg, args...)
}

func (tc *TestCase) LogWarning(warnMsg string, args ...interface{}) {
	tc.warningCount++
	tc.log.LogWarning(warnMsg, args...)
}

func (tc *TestCase) LogPass(passMsg string, args ...interface{}) {
	tc.passedCount++
	tc.log.LogPass(passMsg, args...)
}

func (tc *TestCase) LogMessage(msg string, args ...interface{}) {
	tc.log.LogMessage(msg, args...)
}

func (tc *TestCase) LogDebug(debugMsg string, args ...interface{}) {
	tc.log.LogDebug(debugMsg, args...)
}

func (tc *TestCase) InitParam(name string, value interface{}) interface{} {
	return tc.params.InitParam(name, value)
}

func (tc *TestCase) AddParam(name string, value interface{}, comment string) interface{} {
	return tc.params.AddParam(name, value, comment)
}

func (tc *TestCase) GetParamValue(name string) (interface{}, bool) {
	return tc.params.GetParamValue(name)
}

func (tc *TestCase) GetParams() *Parameters {
	return &tc.params
}

func (tc *TestCase) ReturnFromRun() (int, error) {
	var calcFailThreshold float64
	totalTC := tc.passedCount + tc.failedCount
	if totalTC <= 0 {
		calcFailThreshold = 0
	} else {
		calcFailThreshold = float64((float64(tc.failedCount) / float64(totalTC))) * 100.00
	}

	tc.LogMessage("test %s ran %d check points with failure rate of %.3f", tc.Name(), totalTC, calcFailThreshold)

	if tc.Critical.Triggered() {
		tc.LogError("ERROR:: Found Critical error during run!")
		return TcCriticalError, nil
	}

	if float64(tc.failureThreshold) >= calcFailThreshold {
		return TcPassed, nil
	}
	return TcFailed, nil
}

func (tc *TestCase) RunTest(test iTestCase) {
	chReport := make(chan testResult, 1)
	go tc.parent.Run("", test, chReport)
	_ = <-chReport
}

func InitTest(name string, test iTestCase, parent iTestManager, params Parameters) iTestCase {
	tc := test
	tc.Init(name, parent, params)
	return tc
}
