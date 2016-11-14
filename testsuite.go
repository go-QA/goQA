// Copyright 2013 The goQA Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package goQA

package goQA

import (
	"fmt"
)

type Suite interface {
	Init(name string, parent Manager, params Parameters)
	Setup() (status int, msg string, err error)
	Teardown() (status int, msg string, err error)
	AddTest(test Tester)
	Name() string
	GetTestCase(name string) Tester
	GetTestCases() []Tester
}

type DefaultSuite struct {
	TestCase
	testCases []Tester
}

func (s *DefaultSuite) GetParent() Manager {
	return s.TestCase.parent
}

func (s *DefaultSuite) Init(name string, parent Manager, params Parameters) {
	s.TestCase.Init(name, parent, Parameters{})
	s.testCases = []Tester{}
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

func (s *DefaultSuite) AddTest(test Tester) {
	s.testCases = append(s.testCases, test)
	return
}

func (s *DefaultSuite) RunSuite() {
	s.GetParent().AddSuite(s)
	s.GetParent().RunSuite(s.Name())
}

func (s *DefaultSuite) GetTestCase(name string) Tester {
	for _, test := range s.testCases {
		if test.Name() == name {
			return test
		}
	}
	return nil
}

func (s *DefaultSuite) GetTestCases() []Tester {
	return s.testCases
}

// CreateSuite the default suite object that has basic functionality to run a suite.
func NewSuite(name string, parent Manager, params Parameters) *DefaultSuite {
	suite := DefaultSuite{}
	suite.Init(name, parent, Parameters{})
	return &suite
}
