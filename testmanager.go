// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
	"sync"
	//"error"
	//"os"
	"encoding/xml"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"

	"github.com/go-QA/logger"
)

// control test timing
const (
	MaxTestrunWaitime = 30
)

// --------------------  Default Registery  -----------------------

// TestRegister interface is passed to TestManager in APIs like RunFromXML()
// Used to provide Manager with Test cases and Suites available to From Test Plans
type TestRegister interface {
	GetTestCase(testCaseName string) (Tester, error)
	GetSuite(suiteName string, suiteType string, tm *TestManager, params Parameters) (Suite, error)
}

// DefaultRegister stores TypeOf(<testCase>) in a map:
//  var registry map[string]reflect.Type
//
// Example Of seting up map with two test cases:
//    var regTests map[string]reflect.Type = map[string]reflect.Type{
//        "test1": reflect.TypeOf(Test1{}),
//	      "test2": reflect.TypeOf(Test2{}),
//	      "test3": reflect.TypeOf(Test3{})}
//
type DefaultRegister struct {
	Registry map[string]reflect.Type
}

// GetTestCase Creates the TestCase object and calls Init()
// Tester interface is returned
// error return testError
func (r *DefaultRegister) GetTestCase(testCaseName string) (Tester, error) {

	var test Tester

	if _, ok := r.Registry[testCaseName]; ok {
		test = reflect.New(r.Registry[testCaseName]).Interface().(Tester)
		return test, nil
	}

	return nil, Create(&Parameters{}, "invalid test class '"+testCaseName+"'")
}

// GetSuite Creates DefaultSuite  object and calls Init()
// Suite interface is returned
// error return testError
func (r *DefaultRegister) GetSuite(suiteName string, suiteClass string, tm *TestManager, params Parameters) (Suite, error) {
	suite := DefaultSuite{}
	suite.Init(suiteName, tm, params)
	return &suite, nil
}

// ---------------------------  Define XML for test plans -------------------

// XMLParam can be TestCase, Suite, or Manager parameter
type XMLParam struct {
	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr"`
	Value   string `xml:"value,attr"`
	Comment string `xml:"comment,attr"`
}

// XMLTestCase defines test case
type XMLTestCase struct {
	Name   string     `xml:"name,attr"`
	Class  string     `xml:"class,attr"`
	Params []XMLParam `xml:"Param"`
}

// XMLTestSuite Defines Suite object with list of XMLTestCase and
// suite params
type XMLTestSuite struct {
	Name      string        `xml:"name,attr"`
	Class     string        `xml:"class,attr"`
	Params    []XMLParam    `xml:"Param"`
	TestCases []XMLTestCase `xml:"TestCase"`
}

// XMLTestPlan hold XMLSuite list
type XMLTestPlan struct {
	XMLName xml.Name       `xml:"TestManager"`
	Name    string         `xml:"name,attr"`
	Params  []XMLParam     `xml:"Param"`
	Suites  []XMLTestSuite `xml:"TestSuite"`
}

// --------------------------------------------------------------

type Manager interface {
	RunSuite(suite string) int
	Run(suiteName string, tc Tester, chReport chan testResult)
	AddSuite(suite Suite)
	GetLogger() *logger.GoQALog
	GetSuite(name string) Suite
}

// TestManager runs Suites of tests and uses Reporter interface objects to created
// detailed report(s) for results
type TestManager struct {
	mutex sync.Mutex
	//sMu         sync.Mutex
	//tcStartMu   sync.Mutex
	//tcFinMu     sync.Mutex
	suites     []Suite
	report     ManagerResult
	generators map[string]ReportWriter
	log        *logger.GoQALog
	suiteFlags int
	testFlags  int
}

// Init iTestManager interface method to do setup of Test Manager
func (tm *TestManager) Init(log io.Writer, reportWriter ReportWriter) *TestManager {
	tm.suites = []Suite{}
	tm.report = ManagerResult{}
	tm.generators = make(map[string]ReportWriter)
	tm.report.Init("report1")
	tm.log = &logger.GoQALog{}
	tm.log.Init()
	tm.log.Add("default", logger.LogLevelAll, log)
	//tr := TextReporter{}
	reportWriter.Init(tm)
	tm.addGenerator(reportWriter)
	return tm
}

// GetSuite returns interface Suite based on suite name or nil if not found
func (tm *TestManager) GetSuite(name string) Suite {
	for _, suite := range tm.suites {
		if suite.Name() == name {
			return suite
		}
	}
	return nil
}

// AddLogger creates a new log based on parameters and adds to goQA.goQALog list.
//
func (tm *TestManager) AddLogger(name string, level uint64, stream io.Writer) {
	tm.log.Add(name, level, stream)
}

// Run will execute the TestCase and log test results to chReport
func (tm *TestManager) Run(suiteName string, tc Tester, chReport chan testResult) {
	var (
		runStatus, setupStatus, teardownStatus int
		runErr, setupErr, teardownErr          error
		inRunTest, inRunSetup, inRunTeardown   bool
	)

	inRunSetup = false
	inRunTeardown = false
	inRunTest = false

	result := testResult{}
	result.Init(tc.Name())

	defer func() {
		result.name = tc.Name()
		result.end = time.Now()
		if r := recover(); r != nil {
			result.Status = TcError
			if suiteName != "" {
				if inRunTeardown {
					//fmt.Printf("DEFERING::RECOVER::TEARDOWN=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test Teardown::%s", r)
					result.Status = TcTeardownError
					tm.report.testTeardownError(suiteName, result)
				} else if inRunTest {
					//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test run%s", r)
					tm.report.testError(suiteName, result)
				} else if inRunSetup {
					//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test Setup::%s", r)
					result.Status = TcSetupError
					tm.report.testSetupError(suiteName, result)
				}
			} else {
				result.StatusMessage = "Test complete"
				result.Status = runStatus
			}
		} else {
			result.StatusMessage = "Test complete"
			result.Status = runStatus
		}
		chReport <- result
	}()

	if suiteName != "" {
		tm.report.testStarted(suiteName, tc.Name())
	}

	inRunSetup = true
	setupStatus, setupErr = tc.Setup()
	if setupErr == nil {
		tm.log.LogMessage("TestManager->setup::results=%d", setupStatus)
	}
	//  panic("Panicing in Setup")
	inRunTest = true
	runStatus, runErr = tc.Run()
	if runErr == nil {
		tm.log.LogMessage("TestManager->Run::results=%d", runStatus)
	}

	inRunTeardown = true
	teardownStatus, teardownErr = tc.Teardown()
	if teardownErr == nil {
		tm.log.LogMessage("TestManager->Teardown::results=%d", teardownStatus)
	}
	// panic("Panicing in Teardown")

}

// RunTest is same as Run() but takes Suite name and TestCase name as arguments
func (tm *TestManager) RunTest(suiteName string, testName string) {
	tc := tm.GetSuite(suiteName).GetTestCase(testName)
	tm.Run(suiteName, tc, nil)
}

func (tm *TestManager) RunSuite(suiteName string) int {
	chSuiteResult := make(chan int)
	go tm.runSuite(suiteName, chSuiteResult)
	result := <-chSuiteResult
	return result
}

func (tm *TestManager) runSuite(suiteName string, chSuiteResults chan int) {
	var (
		inSuiteSetup, inSuiteTeardown, inSuiteRuntests bool
		guard                                          chan struct{}
	)

	inSuiteSetup = false
	inSuiteTeardown = false
	chComplete := make(chan int)
	chReport := make(chan testResult)

	suite := tm.GetSuite(suiteName)
	tm.log.LogMessage("Running  Suite '%s'\n", suiteName)

	go tm.endTestHandler(suiteName, chReport)

	defer func() {
		if r := recover(); r != nil {
			if inSuiteTeardown {
				//fmt.Printf("DEFERING::RECOVER::TEARDOWN=%d\n", setupStatus)
				tm.report.suiteTeardownError(suiteName, fmt.Sprintf("Error caught During Suite Teardown::%s", r))
				chSuiteResults <- SuiteTeardownError
			} else if inSuiteRuntests {
				// TODO Need to add cancel runtests
				select {
				case _ = <-chComplete:
					// all ok now
				case <-time.After(time.Second * MaxTestrunWaitime):
					// timed out waiting for runs
				}
				tm.report.suiteFailed(suiteName, fmt.Sprintf("Error caught During Suite Run tests::%s", r))
				chSuiteResults <- SuiteTeardownError

			} else if inSuiteSetup {
				//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
				tm.report.suiteSetupError(suiteName, fmt.Sprintf("Error caught During Suite Setup::%s", r))
				chSuiteResults <- SuiteTeardownError
			} else {
				tm.report.suiteFailed(suiteName, fmt.Sprintf("Error caught During Suite run::%s", r))
				chSuiteResults <- SuiteError
			}
		}
	}()

	tm.report.suiteStarted(suite.Name(), "")

	// Suite Setup
	inSuiteSetup = true
	if status, msg, err := suite.Setup(); err == nil {
		if status == SuiteSetupFailed {
			tm.report.suiteSetupFailed(suite.Name(), msg)
		}
	} else {
		tm.report.suiteSetupError(suite.Name(), err.Error())
	}

	// Run Tests
	inSuiteRuntests = true

	done := make(chan int, 5)
	finished := make(chan int)
	testCount := len(suite.GetTestCases())

	if tm.testFlags != TcSerial {

		if tm.testFlags != TcAll {
			guard = make(chan struct{}, tm.testFlags)
		}

		// Launch a go routine that will increment counter
		// till last test case then senf message on finished channel
		go func() {
			completedTests := 0
			for _ = range done {
				completedTests++
				if completedTests >= testCount {
					break
				}
			}
			finished <- 1
		}()

	}

	for _, tc := range tm.GetSuite(suiteName).GetTestCases() {
		tm.log.LogMessage("Running test '%s'", tc.Name())
		if tm.testFlags == TcAll {
			go tm.launchTest(suiteName, tc, done, chReport)
		} else if tm.testFlags == TcSerial {
			tm.Run(suiteName, tc, chReport)
		} else {
			guard <- struct{}{}
			go tm.launchTestWithGuard(suiteName, tc, guard, done, chReport)
		}
	}

	if tm.testFlags != TcSerial {
		_ = <-finished
	}

	close(chReport)

	// Suite Teardown()
	inSuiteTeardown = true
	if status, msg, err := suite.Teardown(); err == nil {
		if status == SuiteTeardownFailed {
			tm.report.suiteTeardownFailed(suite.Name(), msg)
			chSuiteResults <- SuiteTeardownFailed
		} else {
			tm.report.suitePassed(suite.Name(), "")
			chSuiteResults <- SuitePassed
		}
	} else {
		tm.report.suiteTeardownError(suite.Name(), err.Error())
		chSuiteResults <- SuiteTeardownError
	}
}

func (tm *TestManager) launchTest(suiteName string, testcase Tester, done chan int, chReport chan testResult) {
	tm.Run(suiteName, testcase, chReport)
	done <- 1
}

func (tm *TestManager) launchTestWithGuard(suiteName string, testcase Tester, guard chan struct{}, done chan int, chReport chan testResult) {
	tm.Run(suiteName, testcase, chReport)
	<-guard
	done <- 1
}

func (tm *TestManager) endTestHandler(suiteName string, chResult chan testResult) {
	var result testResult
	//fmt.Printf("LENGTH=%d\n", length)
	for result = range chResult {
		//fmt.Printf("COUNT=%d\n", count)

		switch result.Status {
		case TcNotFound:
			tm.report.testNotFound(suiteName, result)
		case TcSkipped:
			tm.report.testSkipped(suiteName, result)
		case TcPassed:
			tm.report.testPassed(suiteName, result)
		case TcFailed, TcCriticalError:
			tm.report.testFailed(suiteName, result)
		case TcError:
			tm.report.testError(suiteName, result)
		case TcSetupFailed:
			tm.report.testSetupFailed(suiteName, result)
		case TcSetupError:
			tm.report.testSetupError(suiteName, result)
		case TcTeardownFailed:
			tm.report.testTeardownFailed(suiteName, result)
		case TcTeardownError:
			tm.report.testSetupError(suiteName, result)
		}

	}

}

func (tm *TestManager) RunAll() {
	chSuiteResults := make(chan int)
	chComplete := make(chan int)

	tm.report.managerStarted("Test Manager")
	length := len(tm.suites)
	go tm.endManagerHandler(chSuiteResults, chComplete, length)

	tm.log.LogMessage("Running all suitess...")
	if tm.suiteFlags != SuiteAll && tm.suiteFlags != SuiteSerial {
		tm.suiteRunner(chSuiteResults)
	} else {
		for _, suite := range tm.suites {
			if tm.suiteFlags == SuiteAll {
				go tm.runSuite(suite.Name(), chSuiteResults)
			} else if tm.suiteFlags == SuiteSerial {
				tm.runSuite(suite.Name(), chSuiteResults)
			}
		}
	}
	_ = <-chComplete
	tm.report.managerPassed("Test Manager", "")
	tm.managerStatistics("Test Manager", "")
	tm.log.Sync()
}

func (tm *TestManager) convertToParamType(value, paramType string) interface{} {
	var convertedVal interface{}
	switch paramType {
	case "int":
		convertedVal, _ = strconv.ParseInt(value, 10, 64)
	case "float":
		convertedVal, _ = strconv.ParseFloat(value, 64)
	case "string":
		convertedVal = value
	default:
		convertedVal = value
	}
	return convertedVal
}

// ParseTestPlanFromXML tries to read a test plan from XML fileName
// and return *XMLTestPlan.
func (tm *TestManager) ParseTestPlanFromXML(fileName string, testPlan *XMLTestPlan) error {
	buf, err := ioutil.ReadFile(fileName) // "ChamberFunctionality.xml"
	if err != nil {
		fmt.Println(err.Error())
		testPlan = nil
		return err
	}

	tm.log.LogDebug(string(buf))
	err = xml.Unmarshal(buf, &testPlan)
	if err != nil {
		testPlan = nil
	}
	return err
}

// AddTestPlan takes data stored in XMLTestPlan object and adds new suites with tests to Manager
// Suite objects and Test Cases created from TestRegistry interface object
// return nil on success or error
func (tm *TestManager) AddTestPlan(testPlan *XMLTestPlan, registry TestRegister) error {

	var test Tester
	tm.log.LogDebug("%v", testPlan)
	var testParams, suiteParams, MngrParams *Parameters

	MngrParams = new(Parameters)
	MngrParams.Init()
	for _, param := range testPlan.Params {
		MngrParams.AddParam(param.Name, tm.convertToParamType(param.Value, param.Type), param.Comment)
		tm.log.LogDebug("MANAGERPARAM name=%s, type=%s,value= %s, comment=%s", param.Name, param.Type, param.Value, param.Comment)
	}

	for _, xmlSuite := range testPlan.Suites {

		suiteParams = new(Parameters)
		suiteParams.Init()
		for _, param := range xmlSuite.Params {
			suiteParams.AddParam(param.Name, tm.convertToParamType(param.Value, param.Type), param.Comment)
			tm.log.LogDebug("SUITEPARAM name=%s, type=%s,value= %s, comment=%s", param.Name, param.Type, param.Value, param.Comment)
		}
		for k, v := range MngrParams.params {
			if _, ok := suiteParams.params[k]; !ok {
				suiteParams.params[k] = v
			}
		}

		suite, _ := registry.GetSuite(xmlSuite.Name, xmlSuite.Class, tm, *suiteParams)

		for _, xmlTest := range xmlSuite.TestCases {
			testParams = new(Parameters)
			testParams.Init()
			for _, param := range xmlTest.Params {
				testParams.AddParam(param.Name, tm.convertToParamType(param.Value, param.Type), param.Comment)
				tm.log.LogDebug("TESTPARAM name=%s, type=%s,value= %s, comment=%s", param.Name, param.Type, param.Value, param.Comment)
			}
			for k, v := range suiteParams.params {
				if _, ok := testParams.params[k]; !ok {
					testParams.params[k] = v
				}
			}

			test, _ = registry.GetTestCase(xmlTest.Class)
			suite.AddTest(test, xmlTest.Name, *testParams)
		}
		tm.AddSuite(suite)
	}
	return nil
}

// RunFromXML takes XML runplan file runs the suites with test cases by calling RunAll()
func (tm *TestManager) RunFromXML(fileName string, registry TestRegister) error {
	var testPlan XMLTestPlan
	err := tm.ParseTestPlanFromXML(fileName, &testPlan)
	if err != nil {
		tm.log.LogError("Unable to parse XML file %s: error=%s", fileName, err.Error())
		return err
	}

	err = tm.AddTestPlan(&testPlan, registry)
	if err != nil {
		tm.log.LogError("Unable to add test plan::error=%s", err.Error())
		return err
	}
	tm.RunAll()

	return nil
}

func (tm *TestManager) endManagerHandler(chSuiteResult chan int, chComplete chan int, length int) {
	// var result int
	//fmt.Printf("Manager Length = %d\n", length)
	for count := 0; count < length; count++ {
		//fmt.Printf("Manager Count = %d\n", count)
		_ = <-chSuiteResult
	}

	chComplete <- 1
}

func (tm *TestManager) suiteRunner(chSuiteResults chan int) {
	maxSuites := tm.suiteFlags
	guard := make(chan struct{}, maxSuites)

	for _, suite := range tm.suites {
		guard <- struct{}{}
		tm.log.LogMessage("Running Suite '%s'\n", suite.Name())
		go tm.launchSuite(suite.Name(), guard, chSuiteResults)
	}
}

func (tm *TestManager) launchSuite(name string, guard chan struct{}, chSuiteResults chan int) {
	tm.runSuite(name, chSuiteResults)
	<-guard
}

func (tm *TestManager) AddSuite(suite Suite) {
	tm.suites = append(tm.suites, suite)
}

func (tm *TestManager) GetLogger() *logger.GoQALog {
	return tm.log // g.logChannel
}

//  ========= functions to call reporter interface API

func (tm *TestManager) addGenerator(gen ReportWriter) {
	tm.generators[gen.Name()] = gen
}

func (tm *TestManager) managerStatistics(name, msg string) {
	//fmt.Printf("In->managerStatistics()  %p\n", tm.log)
	genCount := len(tm.generators)
	arComplete := make([]chan int, genCount)
	index := 0
	for _, generator := range tm.generators {
		arComplete[index] = make(chan int)
		go generator.PerformManagerStatistics(&tm.report, name, msg, arComplete[index])
		index++
	}
	for i := 0; i < genCount; i++ {
		_ = <-arComplete[i]
	}
}

func NewManager(stream io.Writer, reporter ReportWriter, suiteFlags int, testFlags int) (tm TestManager) {
	tm = TestManager{}
	tm.suiteFlags = suiteFlags
	tm.testFlags = testFlags
	tm.Init(stream, reporter)
	return tm
}
