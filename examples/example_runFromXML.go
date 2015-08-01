package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	//"gitorious.org/goqa/goqa.git"
	"../../goQA"
	"time"
	//"reflect"
	//"net"
	//"encoding/json"
)

const (
	TEST_COUNT = 3
)

type Test1 struct {
	data int
	goQA.TestCase
}

func (t *Test1) Run() (int, error) {
	return 1, nil	
}

type Test2 struct {
	data int
	goQA.TestCase
}

type Test3 struct {
	data int
	goQA.TestCase
}

// interface object 'TestRegister' to create the test cases
type Register int32

func (r *Register) GetTestCase(testName string, tm *goQA.TestManager, params goQA.Parameters) (goQA.ITestCase, error) {
	var test goQA.ITestCase
	switch 	testName {
	case "::ProducerSE::TestCase::Heater::HeaterTest1":
		test = &Test1{}
	case "::ProducerSE::TestCase::Heater::HeaterTest2":
		test = &Test2{}
	default:
		test = &Test1{}
	}

	test.Init(testName, tm, params)
	*r++
	return test, nil
}

func (r *Register) GetSuite(suiteName string, tm *goQA.TestManager, params goQA.Parameters) (goQA.Suite, error) {
	suite := goQA.DefaultSuite{}
	suite.Init(suiteName, tm, params)
	return &suite, nil
}
	

func main() {

	startTime := time.Now()
	//runtime.GOMAXPROCS(2)

	// Report Writer.
	// Only have a TextReporter now that reports plain text to stdout
	tr := goQA.TextReporter{}

	// create the test manager object. Default logger is stdout
	tm := goQA.CreateTestManager(os.Stdout, &tr,
		goQA.SUITE_SERIAL, // Concurency for suites:
		goQA.TC_ALL) // Concurrency for test cases per suite

	console, err := os.Create("data/console.log")
	if err != nil {
		panic(err)
	}
	defer console.Close()
	tm.AddLogger("console", goQA.LOGLEVEL_ALL, console)

	var reg Register
	tm.RunFromXML("ChamberFunctionality.xml", &reg)

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\ntotal runtime  = %.6f\n\n", totalTime)
	//time.Sleep(time.Millisecond * 100)
}
