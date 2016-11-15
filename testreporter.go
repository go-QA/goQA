// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
	//"error"
	//"log"
	//"os"
	//"io"
	"bytes"
	"sync"
	"time"

	"github.com/go-QA/logger"
)

const (
	INITIAL_REPORT_SIZE = 100
)

type ReportWriter interface {
	Name() string
	Init(parent Manager)
	PerformManagerStatistics(report *ManagerResult, name, msg string, complete chan int)
}

// Status codes returned for Test Cases
const (
	_ = iota
	TcNotFound
	TcSkipped
	TcPassed
	TcFailed
	TcError
	TcCriticalError
	TcSetupFailed
	TcSetupError
	TcTeardownFailed
	TcTeardownError
)

// Status codes retuned for suites
const (
	_ = iota
	SuiteOk
	SuiteNotFound
	SuiteSkipped
	SuitePassed
	SuiteFailed
	SuiteError
	SuiteCriticalError
	SuiteSetupFailed
	SuiteSetupError
	SuiteTeardownFailed
	SuiteTeardownError
)

// Status codes returned for manager
const (
	_ = iota
	ManagerPassed
	ManagerFailed
	ManagerSetupFailed
	ManagerSetupError
	ManagerTeardownFailed
	ManagerTeardownError
)

// Text Formating for TextReporter
const (
	TestPassedReport           = "TEST PASSED          %s (%.2f sec) %s"
	TestFailedReport           = "TEST FAILED          %s (%.2f sec) %s"
	TestErrorReport            = "TEST ERROR           %s (%.2f sec) %s"
	TestSetupFailedReport      = "TEST SETUP FAILED    %s %s"
	TestSetupErrorReport       = "TEST SETUP ERROR     %s %s"
	TestTeardownErrorReport    = "TEST TEARDOWN ERROR  %s %s"
	TestNotFondReport          = "TEST NOT FOUND       %s"
	TestSkippedReport          = "TEST SKIPPED         %s"
	SuiteStartedReport         = "SUITE STARTED        %s"
	SuitePassedReport          = "SUITE PASSED         %s (%.2f sec)"
	SuiteFailedReport          = "SUITE FAILED         %s (%.2f sec) %s"
	SuiteErrorReport           = "SUITE ERROR          %s (%.2f sec) %s"
	SuiteSetupFailedReport     = "SUITE SETUP FAILED   %s %s"
	SuiteSetupErrorReport      = "SUITE SETUP ERROR    %s %s"
	SuiteTeardownErrorReport   = "SUITE TEARDOWN ERROR %s %s"
	SuiteNotFoundReport        = "SUITE NOT FOUND      %s"
	ManagerStartedReport       = "MNGR STARTED         %s"
	ManagerSetupFailedReport   = "MNGR SETUP FAILED    %s %s"
	ManagerSetupErrorReport    = "MNGR SETUP ERROR     %s %s"
	ManagerTeardownErrorReport = "MNGR TEARDOWN ERROR  %s %s"
	SuiteStatisticsReport      = "SUITE STATISTICS     %s (%.2f sec)\n\nTests: Total %3d, Passed %3d, Failed %3d, Error %3d, SetUp failed %3d, SetUp error %3d, Not Found %3d"
	ManagerStatisticsReport    = "TOTAL STATISTICS     %s (%.2f sec)\n\nSuites: Total %3d, Passed %3d, Failed %3d, Error %3d, SetUp failed %3d, SetUp error %3d, Not Found %3d\n Tests: Total %3d, Passed %3d, Failed %3d, Error %3d, SetUp failed %3d, SetUp error %3d, Not Found %3d"
)

type ReporterStatistics struct {
	//suites map[string]SuiteStats
	// test manager statistics
	NumberOfTestSuites               int
	NumberOfTestSuitesPassed         int
	NumberOfTestSuitesFailed         int
	NumberOfTestSuitesError          int
	NumberOfTestSuitesSetUpFailed    int
	NumberOfTestSuitesSetUpError     int
	NumberOfTestSuitesTearDownError  int
	NumberOfTestSuitesTearDownFailed int
	NumberOfTestSuitesNotFound       int

	// tm.totals
	TotalNumberOfTestCases               int
	TotalNumberOfTestCasesPassed         int
	TotalNumberOfTestCasesFailed         int
	TotalNumberOfTestCasesError          int
	TotalNumberOfTestCasesSetUpFailed    int
	TotalNumberOfTestCasesSetUpError     int
	TotalNumberOfTestCasesTearDownFailed int
	TotalNumberOfTestCasesTearDownError  int
	TotalNumberOfTestCasesNotFound       int
	TotalNumberOfTestCasesSkipped        int
}

func (s *ReporterStatistics) Init() {
	s.resetManagerResultStatistics()
}

func (s *ReporterStatistics) resetManagerResultStatistics() {
	s.NumberOfTestSuites = 0
	s.NumberOfTestSuitesPassed = 0
	s.NumberOfTestSuitesFailed = 0
	s.NumberOfTestSuitesError = 0
	s.NumberOfTestSuitesSetUpFailed = 0
	s.NumberOfTestSuitesSetUpError = 0
	s.NumberOfTestSuitesTearDownFailed = 0
	s.NumberOfTestSuitesTearDownError = 0
	s.NumberOfTestSuitesNotFound = 0

	s.TotalNumberOfTestCases = 0
	s.TotalNumberOfTestCasesPassed = 0
	s.TotalNumberOfTestCasesFailed = 0
	s.TotalNumberOfTestCasesError = 0
	s.TotalNumberOfTestCasesSetUpFailed = 0
	s.TotalNumberOfTestCasesSetUpError = 0
	s.TotalNumberOfTestCasesTearDownFailed = 0
	s.TotalNumberOfTestCasesTearDownError = 0
	s.TotalNumberOfTestCasesNotFound = 0
	s.TotalNumberOfTestCasesSkipped = 0
}

type suiteResult struct {
	name          string
	Status        int
	start         time.Time
	end           time.Time
	StatusMessage string
	tempTests     map[string]testResult
	tests         []testResult

	// test suite statistics
	NumberOfTestCases               int
	NumberOfTestCasesPassed         int
	NumberOfTestCasesFailed         int
	NumberOfTestCasesError          int
	NumberOfTestCasesSetUpFailed    int
	NumberOfTestCasesSetUpError     int
	NumberOfTestCasesTearDownError  int
	NumberOfTestCasesTearDownFailed int
	NumberOfTestCasesNotFound       int
	NumberOfTestCasesSkipped        int
}

func (s *suiteResult) Init(name string) {
	s.name = name
	s.start = time.Now()
	s.tests = make([]testResult, 0, INITIAL_REPORT_SIZE)
	s.tempTests = make(map[string]testResult)
}

func (s *suiteResult) GetTests() []testResult {
	return s.tests
}

func (s *suiteResult) Runtime() float64 {
	return s.end.Sub(s.start).Seconds()
}

func (s *suiteResult) Name() string {
	return s.name
}

func (s *suiteResult) StartTest(name string) {
	test := testResult{}
	test.Init(name)
	s.tempTests[name] = test
}

func (s *suiteResult) EndTest(name string, status int, message string) {
	test := s.tempTests[name]
	test.Status = status
	test.StatusMessage = message
	test.end = time.Now()
	s.AddTestResult(test)
}

func (s *suiteResult) AddTestResult(result testResult) {
	if len(s.tests) >= cap(s.tests) {
		newSlice := make([]testResult, len(s.tests), len(s.tests)+INITIAL_REPORT_SIZE)
		copy(newSlice, s.tests)
		s.tests = newSlice
	}
	s.tests = append(s.tests, result)

}

type testResult struct {
	name          string
	Status        int
	StatusMessage string
	start         time.Time
	end           time.Time
}

func (t *testResult) Init(name string) {
	t.name = name
	t.start = time.Now()
}

func (t *testResult) Runtime() float64 {
	return t.end.Sub(t.start).Seconds()
}

func (t *testResult) Name() string {
	return t.name
}

type ManagerResult struct {
	mutex          sync.Mutex
	name           string
	start          time.Time
	end            time.Time
	activeSuites   map[string]suiteResult
	finishedSuites []suiteResult
	reportStats    ReporterStatistics
}

func (m *ManagerResult) GetSuites() []suiteResult {
	return m.finishedSuites
}

func (m *ManagerResult) Init(name string) {
	m.name = name
	m.start = time.Now()
	m.activeSuites = make(map[string]suiteResult)
	m.finishedSuites = make([]suiteResult, 0, 10)
}

func (m *ManagerResult) Runtime() float64 {
	return m.end.Sub(m.start).Seconds()
}

func (m *ManagerResult) Name() string {
	return m.name
}

func (m *ManagerResult) StartSuite(name string) {
	//fmt.Printf("ManagerResult::StartSuite::suiteName = %s", name)
	suite := suiteResult{}
	suite.Init(name)
	m.activeSuites[name] = suite
}

func (m *ManagerResult) EndSuite(name string, status int, message string) {
	suite := m.activeSuites[name]
	suite.Status = status
	suite.StatusMessage = message
	suite.end = time.Now()
	m.AddSuiteResult(suite)
}

func (m *ManagerResult) EndManager(name string, status int, message string) {
	m.end = time.Now()
}

func (m *ManagerResult) resetManagerStatistics() {
	m.Init(m.name)
}

func (m *ManagerResult) AddSuiteResult(result suiteResult) {
	if len(m.finishedSuites) >= cap(m.finishedSuites) {
		//fmt.Printf("len=%d   cap=%d\n", len(m.finishedSuites), cap(m.finishedSuites))
		newSlice := make([]suiteResult, len(m.finishedSuites), len(m.finishedSuites)+10)
		copy(newSlice, m.finishedSuites)
		m.finishedSuites = newSlice
	}
	//fmt.Printf("ManagerResult::AddSuiteResult::\n")
	m.finishedSuites = append(m.finishedSuites, result)

}

func (m *ManagerResult) EndTest(suiteName string, result testResult) {
	suite := m.activeSuites[suiteName]
	suite.EndTest(result.name, result.Status, result.StatusMessage)
	m.activeSuites[suiteName] = suite
}

func (m *ManagerResult) StartTest(suiteName string, name string) {
	suite := m.activeSuites[suiteName]
	suite.StartTest(name)
	m.activeSuites[suiteName] = suite
}

func (m *ManagerResult) testStarted(suiteName string, name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.StartTest(suiteName, name)
}

func (m *ManagerResult) testPassed(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesPassed++
	m.reportStats.TotalNumberOfTestCasesPassed++
	m.reportStats.TotalNumberOfTestCases++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testError(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesError++
	m.reportStats.TotalNumberOfTestCasesError++
	m.reportStats.TotalNumberOfTestCases++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testFailed(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesFailed++
	m.reportStats.TotalNumberOfTestCasesFailed++
	m.reportStats.TotalNumberOfTestCases++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)

}

func (m *ManagerResult) testNotFound(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesNotFound++
	m.reportStats.TotalNumberOfTestCases++
	m.reportStats.TotalNumberOfTestCasesNotFound++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testSkipped(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSkipped++
	m.reportStats.TotalNumberOfTestCases++
	m.reportStats.TotalNumberOfTestCasesSkipped++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testSetupFailed(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSetUpFailed++
	m.reportStats.TotalNumberOfTestCases++
	m.reportStats.TotalNumberOfTestCasesSetUpFailed++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testSetupError(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCases++
	s.NumberOfTestCasesSetUpFailed++
	m.reportStats.TotalNumberOfTestCases++
	m.reportStats.TotalNumberOfTestCasesSetUpFailed++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testTeardownFailed(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCasesTearDownFailed++
	m.reportStats.TotalNumberOfTestCasesTearDownFailed++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) testTeardownError(suiteName string, result testResult) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := m.activeSuites[suiteName]
	s.NumberOfTestCasesTearDownError++
	m.reportStats.TotalNumberOfTestCasesTearDownError++
	m.activeSuites[suiteName] = s
	m.EndTest(suiteName, result)
}

func (m *ManagerResult) suitePassed(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reportStats.NumberOfTestSuites++
	m.reportStats.NumberOfTestSuitesPassed++
	m.EndSuite(suiteName, SuitePassed, msg)
}

func (m *ManagerResult) suiteStarted(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	s := suiteResult{}
	m.activeSuites[suiteName] = s
	m.StartSuite(suiteName)
}

func (m *ManagerResult) suiteFailed(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reportStats.NumberOfTestSuites++
	m.reportStats.NumberOfTestSuitesError++
	m.EndSuite(suiteName, SuiteFailed, msg)
}

func (m *ManagerResult) suiteNotFound(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reportStats.NumberOfTestSuites++
	m.reportStats.NumberOfTestSuitesNotFound++
	m.EndSuite(suiteName, SuiteNotFound, msg)
}

func (m *ManagerResult) suiteSkipped(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.EndSuite(suiteName, SuiteSkipped, msg)
	m.EndSuite(suiteName, SuiteSkipped, msg)
}

func (m *ManagerResult) suiteSetupFailed(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reportStats.NumberOfTestSuites++
	m.reportStats.NumberOfTestSuitesSetUpFailed++
}

func (m *ManagerResult) suiteSetupError(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.reportStats.NumberOfTestSuites++
	m.reportStats.NumberOfTestSuitesSetUpError++
}

func (m *ManagerResult) suiteTeardownFailed(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.EndSuite(suiteName, SuiteTeardownFailed, msg)
	m.reportStats.NumberOfTestSuitesTearDownFailed++
}

func (m *ManagerResult) suiteTeardownError(suiteName, msg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.EndSuite(suiteName, SuiteTeardownError, msg)
	m.reportStats.NumberOfTestSuitesTearDownError++
}

func (m *ManagerResult) managerFinished(name string, status int, msg string) {
	m.EndManager(name, status, msg)
}

func (m *ManagerResult) managerPassed(name, msg string) {
	m.managerFinished(name, ManagerPassed, msg)
}

func (m *ManagerResult) managerFailed(name, msg string) {
	m.managerFinished(name, ManagerFailed, msg)
}

func (m *ManagerResult) managerStarted(name string) {
	m.reportStats.Init()
	m.Init(name)
}

func (m *ManagerResult) managerSetUpFailed(name, msg string) {
	m.managerFinished(name, SuiteSetupFailed, msg)
}

func (m *ManagerResult) managerTearDownFailed(name, msg string) {
	m.managerFinished(name, SuiteTeardownFailed, msg)
}

type TextReporter struct {
	name   string
	log    *logger.GoQALog
	parent Manager
	stats  ReporterStatistics
	report ManagerResult
}

func (t *TextReporter) Name() string {
	return t.name
}

func (t *TextReporter) Init(parent Manager) {
	t.report = ManagerResult{}
	t.report.Init("manager special")
	t.name = "TextReporter"
	t.log = parent.GetLogger()
	t.parent = parent

}

func (t *TextReporter) GetSuiteResult() int {
	// TODO calc pass or fail for suite
	return SuitePassed
}

func (t *TextReporter) PerformManagerStatistics(report *ManagerResult, name, msg string, complete chan int) {
	var rep bytes.Buffer
	td := time.Now()
	fmt.Fprintf(&rep, "\n\n")
	fmt.Fprintf(&rep, "-----------------------------------------------------------------------\n\n")
	fmt.Fprintf(&rep, "\n")
	fmt.Fprintf(&rep, "                        Test Run Report '%s'\n\n", name)
	fmt.Fprintf(&rep, "                             %s\n\n", td.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&rep, "\n\n\n")
	fmt.Fprintf(&rep, "            Summary Report:\n")
	fmt.Fprintf(&rep, "\n")
	fmt.Fprintf(&rep, ManagerStatisticsReport, name, report.end.Sub(report.start).Seconds(), report.reportStats.NumberOfTestSuites,
		report.reportStats.NumberOfTestSuitesPassed, report.reportStats.NumberOfTestSuitesFailed, report.reportStats.NumberOfTestSuitesError,
		report.reportStats.NumberOfTestSuitesSetUpFailed, report.reportStats.NumberOfTestSuitesSetUpError,
		report.reportStats.NumberOfTestSuitesNotFound,
		report.reportStats.TotalNumberOfTestCases, report.reportStats.TotalNumberOfTestCasesPassed, report.reportStats.TotalNumberOfTestCasesFailed,
		report.reportStats.TotalNumberOfTestCasesError, report.reportStats.TotalNumberOfTestCasesSetUpFailed,
		report.reportStats.TotalNumberOfTestCasesSetUpError, report.reportStats.TotalNumberOfTestCasesNotFound)

	fmt.Fprintf(&rep, "\n\n\n")
	fmt.Fprintf(&rep, "            Suite Summary:\n")
	for _, suite := range report.finishedSuites {
		fmt.Fprintf(&rep, "\n")
		fmt.Fprintf(&rep, SuiteStatisticsReport, suite.name, suite.end.Sub(suite.start).Seconds(), suite.NumberOfTestCases,
			suite.NumberOfTestCasesPassed, suite.NumberOfTestCasesFailed,
			suite.NumberOfTestCasesError, suite.NumberOfTestCasesSetUpFailed,
			suite.NumberOfTestCasesSetUpError, suite.NumberOfTestCasesNotFound)

		fmt.Fprintf(&rep, "\n")
		switch suite.Status {
		case SuiteOk:
			suiteResult := t.GetSuiteResult()
			if suiteResult == SuitePassed {
				fmt.Fprintf(&rep, SuitePassedReport, suite.name, suite.end.Sub(suite.start).Seconds())
			} else {
				fmt.Fprintf(&rep, SuiteFailedReport, suite.name, suite.end.Sub(suite.start).Seconds(), suite.StatusMessage)
			}
		case SuiteError:
			fmt.Fprintf(&rep, SuiteErrorReport, suite.name, suite.end.Sub(suite.start).Seconds(), suite.StatusMessage)
		case SuiteSetupError:
			fmt.Fprintf(&rep, SuiteSetupErrorReport, suite.name, suite.StatusMessage)
		}

		fmt.Fprintf(&rep, "\n\n")
		fmt.Fprintf(&rep, "               Test Results %s\n", suite.name)
		fmt.Fprintf(&rep, "\n")
		for _, test := range suite.tests {
			switch test.Status {
			case TcPassed:
				fmt.Fprintf(&rep, TestPassedReport, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TcFailed:
				fmt.Fprintf(&rep, TestFailedReport, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TcError:
				fmt.Fprintf(&rep, TestErrorReport, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TcSetupFailed:
				fmt.Fprintf(&rep, TestSetupFailedReport, test.name, test.StatusMessage)
			case TcSetupError:
				fmt.Fprintf(&rep, TestSetupErrorReport, test.name, test.StatusMessage)
			case TcTeardownFailed, TcTeardownError:
				fmt.Fprintf(&rep, TestTeardownErrorReport, test.name, test.StatusMessage)
			case TcSkipped:
				fmt.Fprintf(&rep, TestSkippedReport, test.name)

			}
			fmt.Fprintf(&rep, "\n")
		}
		fmt.Fprintf(&rep, "\n\n")
		fmt.Fprintf(&rep, "-----------------------------------------------------------------------\n\n")

	}
	t.log.LogMessage(rep.String())
	complete <- 1
}
