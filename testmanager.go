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
	"github.com/go-QA/logger"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"time"
)

const (
	MAX_TESTRUN_WAIT = 30
)

// --------------------  Default Registery  -----------------------

// struct with 'goQA.TestRegister' to register the test cases with TestManager
type DefaultRegister struct {
	Registry map[string]reflect.Type
}

func (r *DefaultRegister) GetTestCase(testName string, testClass string, tm *TestManager, params Parameters) (ITestCase, error) {

	var test ITestCase

	if _, ok := r.Registry[testClass]; ok {
		test = reflect.New(r.Registry[testClass]).Interface().(ITestCase)
		test.Init(testName, tm, params)
		return test, nil
	}

	return nil, Create(&Parameters{}, "invalid test class '"+testName+"'")
}

func (r *DefaultRegister) GetSuite(suiteName string, suiteClass string, tm *TestManager, params Parameters) (Suite, error) {
	suite := DefaultSuite{}
	suite.Init(suiteName, tm, params)
	return &suite, nil
}

// --------------------------------------------------------------------

// ---------------------------  Define XML for test plans -------------------

type TestRegister interface {
	GetTestCase(testName string, testType string, tm *TestManager, params Parameters) (ITestCase, error)
	GetSuite(suiteName string, suiteType string, tm *TestManager, params Parameters) (Suite, error)
}

type XMLParam struct {
	Name    string `xml:"name,attr"`
	Type    string `xml:"type,attr"`
	Value   string `xml:"value,attr"`
	Comment string `xml:"comment,attr"`
}

type XMLTestCase struct {
	Name   string     `xml:"name,attr"`
	Class  string     `xml:"class,attr"`
	Params []XMLParam `xml:"Param"`
}

type XMLTestSuite struct {
	Name      string        `xml:"name,attr"`
	Class     string        `xml:"class,attr"`
	Params    []XMLParam    `xml:"Param"`
	TestCases []XMLTestCase `xml:"TestCase"`
}

type XMLTestPlan struct {
	XMLName xml.Name       `xml:"TestManager"`
	Name    string         `xml:"name,attr"`
	Suites  []XMLTestSuite `xml:"TestSuite"`
}

// --------------------------------------------------------------

type iTestManager interface {
	RunSuite(suite string, chSuiteResults chan int)
	Run(suiteName string, tc iTestCase, chReport chan testResult)
	AddSuite(suite Suite)
	GetLogger() *logger.GoQALog
	GetSuite(name string) Suite
}

type TestManager struct {
	mutex       sync.Mutex
	sMu         sync.Mutex
	tcStartMu   sync.Mutex
	tcFinMu     sync.Mutex
	suites      []Suite
	reportStats ReporterStatistics
	report      ManagerResult
	generators  map[string]ReportWriter
	log         *logger.GoQALog
	suiteFlags  int
	testFlags   int
}

func (tm *TestManager) Init(log io.Writer, reportWriter ReportWriter) *TestManager {
	tm.suites = []Suite{}
	tm.reportStats = ReporterStatistics{}
	tm.report = ManagerResult{}
	tm.generators = make(map[string]ReportWriter)
	tm.reportStats.Init()
	tm.report.Init("report1")
	tm.log = &logger.GoQALog{}
	tm.log.Init()
	tm.log.Add("default", logger.LOGLEVEL_ALL, log)
	//tr := TextReporter{}
	reportWriter.Init(tm)
	tm.addGenerator(reportWriter)
	return tm
}

func (tm *TestManager) GetSuite(name string) Suite {
	for _, suite := range tm.suites {
		if suite.Name() == name {
			return suite
		}
	}
	return nil
}

func (tm *TestManager) AddLogger(name string, level uint64, stream io.Writer) {
	tm.log.Add(name, level, stream)
}

func (tm *TestManager) Run(suiteName string, tc iTestCase, chReport chan testResult) {
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
			result.Status = TC_ERROR
			if suiteName != "" {
				if inRunTeardown {
					//fmt.Printf("DEFERING::RECOVER::TEARDOWN=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test Teardown::%s", r)
					result.Status = TC_TEARDOWN_ERROR
					tm.testTeardownError(suiteName, result)
				} else if inRunTest {
					//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test run%s", r)
					tm.testError(suiteName, result)
				} else if inRunSetup {
					//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
					result.StatusMessage = fmt.Sprintf("Error caught During test Setup::%s", r)
					result.Status = TC_SETUP_ERROR
					tm.testSetupError(suiteName, result)
				}
			} else {
				result.StatusMessage = fmt.Sprintf("Error caught During test run%s", r)
			}
		} else {
			result.StatusMessage = "Test complete"
			result.Status = runStatus
		}
		chReport <- result
	}()

	if suiteName != "" {
		tm.testStarted(suiteName, tc.Name())
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

func (tm *TestManager) RunTest(suiteName string, testName string) {
	tc := tm.GetSuite(suiteName).GetTestCase(testName)
	tm.Run(suiteName, tc, nil)
}

func (tm *TestManager) RunSuite(suiteName string, chSuiteResults chan int) {
	var (
		// runStatus, setupStatus, teardownStatus int
		//runErr, setupErr, teardownErr          error
		inSuiteSetup, inSuiteTeardown, inSuiteRuntests bool
	)

	inSuiteSetup = false
	inSuiteTeardown = false
	chComplete := make(chan int)
	chReport := make(chan testResult)

	suite := tm.GetSuite(suiteName)
	tm.log.LogMessage("Running  Suite '%s'\n", suiteName)

	length := len(suite.GetTestCases())
	go tm.endSuiteHandler(suiteName, chReport, chComplete, length)

	defer func() {
		if r := recover(); r != nil {
			if inSuiteTeardown {
				//fmt.Printf("DEFERING::RECOVER::TEARDOWN=%d\n", setupStatus)
				tm.suiteTeardownError(suiteName, fmt.Sprintf("Error caught During Suite Teardown::%s", r))
				chSuiteResults <- SUITE_TEARDOWN_ERROR
			} else if inSuiteRuntests {
				// TODO Need to add cancel runtests
				select {
				case _ = <-chComplete:
					// all ok now
				case <-time.After(time.Second * MAX_TESTRUN_WAIT):
					// timed out waiting for runs
				}
				tm.suiteFailed(suiteName, fmt.Sprintf("Error caught During Suite Run tests::%s", r))
				chSuiteResults <- SUITE_TEARDOWN_ERROR

			} else if inSuiteSetup {
				//fmt.Printf("DEFERING::RECOVER::SETUP=%d\n", setupStatus)
				tm.suiteSetupError(suiteName, fmt.Sprintf("Error caught During Suite Setup::%s", r))
				chSuiteResults <- SUITE_TEARDOWN_ERROR
			} else {
				tm.suiteFailed(suiteName, fmt.Sprintf("Error caught During Suite run::%s", r))
				chSuiteResults <- SUITE_ERROR
			}
		}
	}()

	tm.suiteStarted(suite.Name(), "")

	// Suite Setup
	inSuiteSetup = true
	if status, msg, err := suite.Setup(); err == nil {
		if status == SUITE_SETUP_FAILED {
			tm.suiteSetupFailed(suite.Name(), msg)
		}
	} else {
		tm.suiteSetupError(suite.Name(), err.Error())
	}

	// Run Tests
	inSuiteRuntests = true
	if tm.testFlags != TC_ALL && tm.testFlags != TC_SERIAL {
		tm.tcRunner(suiteName, chReport)
	} else {
		count := 0
		for _, tc := range tm.GetSuite(suiteName).GetTestCases() {
			count++
			if count >= length {
				//fmt.Println("LAUNCHING LAST TEST")
			}
			tm.log.LogMessage("Running test '%s'", tc.Name())
			if tm.testFlags == TC_ALL {
				go tm.Run(suiteName, tc, chReport)
			} else {
				tm.Run(suiteName, tc, chReport)
			}
		}
		if tm.testFlags == TC_ALL {
		}
	}

	// wait for all tests complete
	_ = <-chComplete
	// Suite Teardown()
	inSuiteTeardown = true
	if status, msg, err := suite.Teardown(); err == nil {
		if status == SUITE_TEARDOWN_FAILED {
			tm.suiteTeardownFailed(suite.Name(), msg)
			chSuiteResults <- SUITE_TEARDOWN_FAILED
		} else {
			tm.suitePassed(suite.Name(), "")
			chSuiteResults <- SUITE_PASSED
		}
	} else {
		tm.suiteTeardownError(suite.Name(), err.Error())
		chSuiteResults <- SUITE_TEARDOWN_ERROR
	}
}

func (tm *TestManager) endSuiteHandler(suiteName string, chResult chan testResult, chComplete chan int, length int) {
	var result testResult
	//fmt.Printf("LENGTH=%d\n", length)
	for count := 0; count < length; count++ {
		//fmt.Printf("COUNT=%d\n", count)
		result = <-chResult

		switch result.Status {
		case TC_NOT_FOUND:
			tm.testNotFound(suiteName, result)
		case TC_SKIPPED:
			tm.testSkipped(suiteName, result)
		case TC_PASSED:
			tm.testPassed(suiteName, result)
		case TC_FAILED:
			tm.testFailed(suiteName, result)
		case TC_ERROR:
			tm.testError(suiteName, result)
		case TC_SETUP_FAILED:
			tm.testSetupFailed(suiteName, result)
		case TC_SETUP_ERROR:
			tm.testSetupError(suiteName, result)
		case TC_TEARDOWN_FAILED:
			tm.testTeardownFailed(suiteName, result)
		case TC_TEARDOWN_ERROR:
			tm.testSetupError(suiteName, result)
		}

		tm.mutex.Lock()
		tm.report.EndTest(suiteName, result)
		tm.mutex.Unlock()
	}

	chComplete <- 1
}

func (tm *TestManager) RunAll() {
	chSuiteResults := make(chan int)
	chComplete := make(chan int)

	tm.managerStarted("Test Manager")
	length := len(tm.suites)
	go tm.endManagerHandler(chSuiteResults, chComplete, length)

	tm.log.LogMessage("Running all suitess...")
	if tm.suiteFlags != SUITE_ALL && tm.suiteFlags != SUITE_SERIAL {
		tm.suiteRunner(chSuiteResults)
	} else {
		for _, suite := range tm.suites {
			if tm.suiteFlags == SUITE_ALL {
				go tm.RunSuite(suite.Name(), chSuiteResults)
			} else if tm.suiteFlags == SUITE_SERIAL {
				tm.RunSuite(suite.Name(), chSuiteResults)
			}
		}
	}
	_ = <-chComplete
	tm.managerPassed("Test Manager", "")
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

func (tm *TestManager) RunFromXML(fileName string, registry TestRegister) {

	buf, err := ioutil.ReadFile(fileName) // "ChamberFunctionality.xml"
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var testPlan XMLTestPlan
	var params *Parameters
	var test iTestCase

	tm.log.LogDebug(string(buf))
	xml.Unmarshal(buf, &testPlan)

	tm.log.LogDebug("%v", testPlan)
	for _, xmlSuite := range testPlan.Suites {
		suite, _ := registry.GetSuite(xmlSuite.Name, xmlSuite.Class, tm, Parameters{})
		for _, xmlTest := range xmlSuite.TestCases {

			params = new(Parameters)
			for _, param := range xmlTest.Params {
				params.AddParam(param.Name, tm.convertToParamType(param.Value, param.Type), param.Comment)
				tm.log.LogDebug("name=%s, type=%s,value= %s, comment=%s", param.Name, param.Type, param.Value, param.Comment)
			}
			test, _ = registry.GetTestCase(xmlTest.Name, xmlTest.Class, tm, *params)
			suite.AddTest(test)
		}
		tm.AddSuite(suite)
	}
	tm.RunAll()
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
	done := make(chan int, 5)
	finished := make(chan int)
	suiteCount := len(tm.suites)

	go func() {
		completedSuites := 0
		for completedSuites < suiteCount {
			_ = <-done
			completedSuites++
		}
		finished <- 1
	}()

	for _, suite := range tm.suites {
		guard <- struct{}{}
		tm.log.LogMessage("Running Suite '%s'\n", suite.Name())
		go tm.launchSuite(suite.Name(), guard, done, chSuiteResults)
	}

	// wait for all remaining tests before exiting
	_ = <-finished
}

func (tm *TestManager) launchSuite(name string, guard chan struct{}, done chan int, chSuiteResults chan int) {
	tm.RunSuite(name, chSuiteResults)
	<-guard
	done <- 1
}

func (tm *TestManager) tcRunner(suiteName string, chReport chan testResult) {
	maxTests := tm.testFlags
	guard := make(chan struct{}, maxTests)
	done := make(chan int, 5)
	finished := make(chan int)
	suite := tm.GetSuite(suiteName)
	testCount := len(suite.GetTestCases())

	go func() {
		completedTests := 0
		for completedTests < testCount {
			_ = <-done
			completedTests++
		}
		finished <- 1
	}()

	for _, tc := range suite.GetTestCases() {
		guard <- struct{}{}
		tm.log.LogMessage("Running test '%s'\n", tc.Name())
		go tm.launchTest(suiteName, tc, guard, done, chReport)
	}

	// wait for all remaining tests before exiting
	_ = <-finished
}

func (tm *TestManager) launchTest(suiteName string, testcase iTestCase, guard chan struct{}, done chan int, chReport chan testResult) {
	tm.Run(suiteName, testcase, chReport)
	<-guard
	done <- 1
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

func (tm *TestManager) testStarted(suiteName string, name string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.StartTest(suiteName, name)
}

func (tm *TestManager) testPassed(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesPassed++
	tm.reportStats.TotalNumberOfTestCasesPassed++
	tm.reportStats.TotalNumberOfTestCases++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testError(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesError++
	tm.reportStats.TotalNumberOfTestCasesError++
	tm.reportStats.TotalNumberOfTestCases++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testFailed(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesFailed++
	tm.reportStats.TotalNumberOfTestCasesFailed++
	tm.reportStats.TotalNumberOfTestCases++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)

}

func (tm *TestManager) testNotFound(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesNotFound++
	tm.reportStats.TotalNumberOfTestCases++
	tm.reportStats.TotalNumberOfTestCasesNotFound++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testSkipped(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSkipped++
	tm.reportStats.TotalNumberOfTestCases++
	tm.reportStats.TotalNumberOfTestCasesSkipped++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testSetupFailed(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSetUpFailed++
	tm.reportStats.TotalNumberOfTestCases++
	tm.reportStats.TotalNumberOfTestCasesSetUpFailed++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testSetupError(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSetUpFailed++
	tm.reportStats.TotalNumberOfTestCases++
	tm.reportStats.TotalNumberOfTestCasesSetUpFailed++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testTeardownFailed(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCasesTearDownFailed++
	tm.reportStats.TotalNumberOfTestCasesTearDownFailed++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) testTeardownError(suiteName string, result testResult) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := tm.report.tempSuites[suiteName]
	s.NumberOfTestCasesTearDownError++
	tm.reportStats.TotalNumberOfTestCasesTearDownError++
	tm.report.tempSuites[suiteName] = s
	tm.report.EndTest(suiteName, result)
}

func (tm *TestManager) suitePassed(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_PASSED, msg)
	tm.reportStats.NumberOfTestSuites++
	tm.reportStats.NumberOfTestSuitesPassed++
}

func (tm *TestManager) suiteStarted(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	s := suiteResult{}
	tm.report.tempSuites[suiteName] = s
	tm.report.StartSuite(suiteName)
}

func (tm *TestManager) suiteFailed(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_FAILED, msg)
	tm.reportStats.NumberOfTestSuites++
	tm.reportStats.NumberOfTestSuitesError++
}

func (tm *TestManager) suiteNotFound(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_NOT_FOUND, msg)
	tm.reportStats.NumberOfTestSuites++
	tm.reportStats.NumberOfTestSuitesNotFound++
}

func (tm *TestManager) suiteSkipped(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_SKIPPED, msg)
}

func (tm *TestManager) suiteSetupFailed(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.reportStats.NumberOfTestSuites++
	tm.reportStats.NumberOfTestSuitesSetUpFailed++
}

func (tm *TestManager) suiteSetupError(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.reportStats.NumberOfTestSuites++
	tm.reportStats.NumberOfTestSuitesSetUpError++
}

func (tm *TestManager) suiteTeardownFailed(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_TEARDOWN_FAILED, msg)
	tm.reportStats.NumberOfTestSuitesTearDownFailed++
}

func (tm *TestManager) suiteTeardownError(suiteName, msg string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.report.EndSuite(suiteName, SUITE_TEARDOWN_ERROR, msg)
	tm.reportStats.NumberOfTestSuitesTearDownError++
}

func (tm *TestManager) managerFinished(name string, status int, msg string) {
	tm.report.EndManager(name, status, msg)
	tm.managerStatistics(name, msg)
}

func (tm *TestManager) managerPassed(name, msg string) {
	tm.managerFinished(name, MANAGER_PASSED, msg)
}

func (tm *TestManager) managerFailed(name, msg string) {
	tm.managerFinished(name, MANAGER_FAILED, msg)
}

func (tm *TestManager) managerStarted(name string) {
	tm.reportStats.Init()
	tm.report.Init(name)
}

func (tm *TestManager) managerSetUpFailed(name, msg string) {
	tm.managerFinished(name, SUITE_SETUP_FAILED, msg)
}

func (tm *TestManager) managerTearDownFailed(name, msg string) {
	tm.managerFinished(name, SUITE_TEARDOWN_FAILED, msg)
}

func (tm *TestManager) managerStatistics(name, msg string) {
	//fmt.Printf("In->managerStatistics()  %p\n", tm.log)
	genCount := len(tm.generators)
	arComplete := make([]chan int, genCount)
	index := 0
	for _, generator := range tm.generators {
		arComplete[index] = make(chan int)
		go generator.PerformManagerStatistics(&tm.report, &tm.reportStats, name, msg, arComplete[index])
		index++
	}
	for i := 0; i < genCount; i++ {
		_ = <-arComplete[i]
	}
}

func CreateTestManager(stream io.Writer, reporter ReportWriter, suiteFlags int, testFlags int) (tm TestManager) {
	tm = TestManager{}
	tm.suiteFlags = suiteFlags
	tm.testFlags = testFlags
	tm.Init(stream, reporter)
	return tm
}
