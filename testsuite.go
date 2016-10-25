// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
)

type Suite interface {
	//Init(name string, parent iTestManager) Suite
	Init(name string, parent iTestManager, params Parameters)
	Setup() (status int, msg string, err error)
	Teardown() (status int, msg string, err error)
	AddTest(test iTestCase)
	Name() string
	GetTestCase(name string) iTestCase
	GetTestCases() []iTestCase
}

type DefaultSuite struct {
	TestCase
	testCases []iTestCase
}

func (s *DefaultSuite) GetParent() iTestManager {
	return s.TestCase.parent
}

func (s *DefaultSuite) Init(name string, parent iTestManager, params Parameters) {
	s.TestCase.Init(name, parent, Parameters{})
	s.testCases = []iTestCase{}
}

func (s *DefaultSuite) Name() string {
	return s.TestCase.name
}

func (s *DefaultSuite) Setup() (status int, msg string, err error) {
	// Suite setup
	fMsg := fmt.Sprintf("SUITE(%s) SETUP::Status = %d::Message=%s", s.Name(), status, msg)
	s.LogMessage(fMsg)
	//g.logger.Printf("PASS::%s\n", passMsg)
	return SuiteOk, "", nil
}

func (s *DefaultSuite) Teardown() (status int, msg string, err error) {
	// Suite teardown
	fMsg := fmt.Sprintf("SUITE(%s) TEARDOWN::Status = %d::Message=%s", s.Name(), status, msg)
	s.LogMessage(fMsg)
	return SuiteOk, "", nil
}

func (s *DefaultSuite) AddTest(test iTestCase) {
	s.testCases = append(s.testCases, test)
	return
}

func (s *DefaultSuite) RunSuite() {
	chSuite := make(chan int)
	s.GetParent().AddSuite(s)
	go s.GetParent().RunSuite(s.Name(), chSuite)
	_ = <-chSuite
}

func (s *DefaultSuite) GetTestCase(name string) iTestCase {
	for _, test := range s.testCases {
		if test.Name() == name {
			return test
		}
	}
	return nil
}

func (s *DefaultSuite) GetTestCases() []iTestCase {
	return s.testCases
}

// create the default suite object that has basic functionality to run a suite.
func CreateSuite(name string, parent iTestManager, params Parameters) *DefaultSuite {
	suite := DefaultSuite{}
	suite.Init(name, parent, Parameters{})
	return &suite
}
