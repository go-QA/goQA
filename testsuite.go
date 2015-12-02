// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
)

type Suite interface {
	//Init(name string, parent iTestManager) Suite
	Setup() (status int, msg string, err error)
	Teardown() (status int, msg string, err error)
	AddTest(test iTestCase)
	Name() string
	GetTestCase(name string) iTestCase
	GetTestCases() map[string]iTestCase
}

type DefaultSuite struct {
	TestCase
	testCases map[string]iTestCase
}

func (s *DefaultSuite) Init(name string, parent iTestManager, params Parameters) {
	s.testCases = make(map[string]iTestCase)
	s.TestCase.Init(name, parent, Parameters{})
	s.TestCase.name = name
	s.TestCase.log = parent.GetLogger()
	s.TestCase.parent = parent
}

func (s *DefaultSuite) Name() string {
	return s.TestCase.name
}

func (s *DefaultSuite) Setup() (status int, msg string, err error) {
	// Suite setup
	fMsg := fmt.Sprintf("SUITE(%s) SETUP::Status = %d::Message=%s", s.TestCase.name, status, msg)
	s.TestCase.LogMessage(fMsg)
	//g.logger.Printf("PASS::%s\n", passMsg)
	return SUITE_OK, "", nil
}

func (s *DefaultSuite) Teardown() (status int, msg string, err error) {
	// Suite teardown
	fMsg := fmt.Sprintf("SUITE(%s) TEARDOWN::Status = %d::Message=%s", s.TestCase.name, status, msg)
	s.TestCase.LogMessage(fMsg)
	return SUITE_OK, "", nil
}

func (s *DefaultSuite) AddTest(test iTestCase) {
	newTest := test
	s.testCases[test.Name()] = newTest
}

func (s *DefaultSuite) RunSuite() {
	chSuite := make(chan int)
	s.TestCase.parent.AddSuite(s)
	go s.TestCase.parent.RunSuite(s.TestCase.name, chSuite)
	_ = <-chSuite
}

func (s *DefaultSuite) GetTestCase(name string) iTestCase {
	return s.testCases[name]
}

func (s *DefaultSuite) GetTestCases() map[string]iTestCase {
	return s.testCases
}

// create the default suite object that has basic functionality to run a suite.
func CreateSuite(name string, parent iTestManager, params Parameters) *DefaultSuite {
	suite := DefaultSuite{}
	suite.Init(name, parent, Parameters{})
	return &suite
}
