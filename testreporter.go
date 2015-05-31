// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	//"fmt"
	//"error"
	//"log"
	//"os"
	//"io"
	"time"
)

const (
	INITIAL_REPORT_SIZE = 100
)

type ReportWriter interface {
	Name() string
	Init(parent ITestManager)
	PerformManagerStatistics(report *ManagerResult, stats *ReporterStatistics, name, msg string, complete chan int)
}

const (
	_ = iota
	TC_NOT_FOUND
	TC_SKIPPED
	TC_PASSED
	TC_FAILED
	TC_ERROR
	TC_SETUP_FAILED
	TC_SETUP_ERROR
	TC_TEARDOWN_FAILED
	TC_TEARDOWN_ERROR
)

const (
	_ = iota
	SUITE_OK
	SUITE_NOT_FOUND
	SUITE_SKIPPED
	SUITE_PASSED
	SUITE_FAILED
	SUITE_ERROR
	SUITE_SETUP_FAILED
	SUITE_SETUP_ERROR
	SUITE_TEARDOWN_FAILED
	SUITE_TEARDOWN_ERROR
)

const (
	_ = iota
	MANAGER_PASSED
	MANAGER_FAILED
	MANAGER_SETUP_FAILED
	MANAGER_SETUP_ERROR
	MANAGER_TEARDOWN_FAILED
	MANAGER_TEARDOWN_ERROR
)

const (
	TEST_PASSED_REPORT            = "TEST PASSED          %s (%.2f sec) %s"
	TEST_FAILED_REPORT            = "TEST FAILED          %s (%.2f sec) %s"
	TEST_ERROR_REPORT             = "TEST ERROR           %s (%.2f sec) %s"
	TEST_SETUP_FAILED_REPORT      = "TEST SETUP FAILED    %s %s"
	TEST_SETUP_ERROR_REPORT       = "TEST SETUP ERROR     %s %s"
	TEST_TEARDOWN_ERROR_REPORT    = "TEST TEARDOWN ERROR  %s %s"
	TEST_NOT_FOUND_REPORT         = "TEST NOT FOUND       %s"
	TEST_SKIPPED_REPORT           = "TEST SKIPPED         %s"
	SUITE_STARTED_REPORT          = "SUITE STARTED        %s"
	SUITE_PASSED_REPORT           = "SUITE PASSED         %s (%.2f sec)"
	SUITE_FAILED_REPORT           = "SUITE FAILED         %s (%.2f sec) %s"
	SUITE_ERROR_REPORT            = "SUITE ERROR          %s (%.2f sec) %s"
	SUITE_SETUP_FAILED_REPORT     = "SUITE SETUP FAILED   %s %s"
	SUITE_SETUP_ERROR_REPORT      = "SUITE SETUP ERROR    %s %s"
	SUITE_TEARDOWN_ERROR_REPORT   = "SUITE TEARDOWN ERROR %s %s"
	SUITE_NOT_FOUND_REPORT        = "SUITE NOT FOUND      %s"
	MANAGER_STARTED_REPORT        = "MNGR STARTED         %s"
	MANAGER_SETUP_FAILED_REPORT   = "MNGR SETUP FAILED    %s %s"
	MANAGER_SETUP_ERROR_REPORT    = "MNGR SETUP ERROR     %s %s"
	MANAGER_TEARDOWN_ERROR_REPORT = "MNGR TEARDOWN ERROR  %s %s"
	SUITE_STATISTICS_REPORT       = "SUITE STATISTICS     %s (%.2f sec)\nTests: Total %d, Passed %d, Failed %d, Error %d, SetUp failed %d, SetUp error %d, Not Found %d"
	MANAGER_STATISTICS_REPORT     = "TOTAL STATISTICS     %s (%.2f sec)\nSuites: Total %d, Passed %d, Failed %d, Error %d, SetUp failed %d, SetUp error %d, Not Found %d\nTests: Total %d, Passed %d, Failed %d, Error %d, SetUp failed %d, SetUp error %d, Not Found %d"
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
	s.resetTestManagerStatistics()
}

func (s *ReporterStatistics) resetTestManagerStatistics() {
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

type ManagerResult struct {
	name       string
	start      time.Time
	end        time.Time
	tempSuites map[string]suiteResult
	suites     []suiteResult
}

func (m *ManagerResult) GetSuites() []suiteResult {
	return m.suites
}

func (m *ManagerResult) Init(name string) {
	m.name = name
	m.start = time.Now()
	m.tempSuites = make(map[string]suiteResult)
	m.suites = make([]suiteResult, 0, 10)
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
	m.tempSuites[name] = suite
}

func (m *ManagerResult) EndSuite(name string, status int, message string) {
	suite := m.tempSuites[name]
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
	if len(m.suites) >= cap(m.suites) {
		//fmt.Printf("len=%d   cap=%d\n", len(m.suites), cap(m.suites))
		newSlice := make([]suiteResult, len(m.suites), len(m.suites)+10)
		copy(newSlice, m.suites)
		m.suites = newSlice
	}
	//fmt.Printf("ManagerResult::AddSuiteResult::\n")
	m.suites = append(m.suites, result)

}

func (m *ManagerResult) EndTest(suiteName string, result testResult) {
	suite := m.tempSuites[suiteName]
	suite.EndTest(result.name, result.Status, result.StatusMessage)
	m.tempSuites[suiteName] = suite
}

func (m *ManagerResult) StartTest(suiteName string, name string) {
	suite := m.tempSuites[suiteName]
	suite.StartTest(name)
	m.tempSuites[suiteName] = suite
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

type TextReporter struct {
	name   string
	logger *GoQALog
	parent ITestManager
	stats  ReporterStatistics
	report ManagerResult
}

func (t *TextReporter) Name() string {
	return t.name
}

func (t *TextReporter) Init(parent ITestManager) {
	t.report = ManagerResult{}
	t.report.Init("manager special")
	t.name = "TextReporter"
	t.logger = parent.GetLogger()
	t.parent = parent

}

func (t *TextReporter) GetSuiteResult() int {
	// TODO calc pass or fail for suite
	return SUITE_PASSED
}

func (t *TextReporter) PerformManagerStatistics(report *ManagerResult, stats *ReporterStatistics, name, msg string, complete chan int) {
	t.logger.LogMessage("\n\n")
	t.logger.LogMessage(MANAGER_STATISTICS_REPORT, name, report.end.Sub(report.start).Seconds(), stats.NumberOfTestSuites,
		stats.NumberOfTestSuitesPassed, stats.NumberOfTestSuitesFailed, stats.NumberOfTestSuitesError,
		stats.NumberOfTestSuitesSetUpFailed, stats.NumberOfTestSuitesSetUpError,
		stats.NumberOfTestSuitesNotFound,
		stats.TotalNumberOfTestCases, stats.TotalNumberOfTestCasesPassed, stats.TotalNumberOfTestCasesFailed,
		stats.TotalNumberOfTestCasesError, stats.TotalNumberOfTestCasesSetUpFailed,
		stats.TotalNumberOfTestCasesSetUpError, stats.TotalNumberOfTestCasesNotFound)

	t.logger.LogMessage("\n\n")
	for _, suite := range report.suites {
		t.logger.LogMessage("\n\n")
		t.logger.LogMessage(SUITE_STATISTICS_REPORT, suite.name, suite.end.Sub(suite.start).Seconds(), suite.NumberOfTestCases,
			suite.NumberOfTestCasesPassed, suite.NumberOfTestCasesFailed,
			suite.NumberOfTestCasesError, suite.NumberOfTestCasesSetUpFailed,
			suite.NumberOfTestCasesSetUpError, suite.NumberOfTestCasesNotFound)

		t.logger.LogMessage("\n")
		switch suite.Status {
		case SUITE_OK:
			suiteResult := t.GetSuiteResult()
			if suiteResult == SUITE_PASSED {
				t.logger.LogPass(SUITE_PASSED_REPORT, suite.name, suite.end.Sub(suite.start).Seconds())
			} else {
				t.logger.LogFail(SUITE_FAILED_REPORT, suite.name, suite.end.Sub(suite.start).Seconds(), suite.StatusMessage)
			}
		case SUITE_ERROR:
			t.logger.LogMessage(SUITE_ERROR_REPORT, suite.name, suite.end.Sub(suite.start).Seconds(), suite.StatusMessage)
		case SUITE_SETUP_ERROR:
			t.logger.LogMessage(SUITE_SETUP_ERROR_REPORT, suite.name, suite.StatusMessage)
		}

		t.logger.LogMessage("\n")
		for _, test := range suite.tests {
			switch test.Status {
			case TC_PASSED:
				t.logger.LogPass(TEST_PASSED_REPORT, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TC_FAILED:
				t.logger.LogFail(TEST_FAILED_REPORT, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TC_ERROR:
				t.logger.LogError(TEST_ERROR_REPORT, test.name, test.end.Sub(test.start).Seconds(), test.StatusMessage)
			case TC_SETUP_FAILED:
				t.logger.LogMessage(TEST_SETUP_FAILED_REPORT, test.name, test.StatusMessage)
			case TC_SETUP_ERROR:
				t.logger.LogMessage(TEST_SETUP_ERROR_REPORT, test.name, test.StatusMessage)
			case TC_TEARDOWN_FAILED, TC_TEARDOWN_ERROR:
				t.logger.LogMessage(TEST_TEARDOWN_ERROR_REPORT, test.name, test.StatusMessage)
			case TC_SKIPPED:
				t.logger.LogMessage(TEST_SKIPPED_REPORT, test.name)

			}
		}

	}
	complete <- 1
}
