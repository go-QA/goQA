package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	"github.com/go-QA/goQA"
	"github.com/go-QA/logger"
	"time"
	"reflect"
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

var regTests map[string]reflect.Type = map[string]reflect.Type{ "test1": reflect.TypeOf(Test1{}),
																"test2": reflect.TypeOf(Test3{}),
																"test3": reflect.TypeOf(Test3{})}

// struct with 'goQA.TestRegister' to register the test cases with TestManager
type Register struct {
	Registry map[string]reflect.Type
}

func (r *Register) GetTestCase(testName string, tm *goQA.TestManager, params goQA.Parameters) (goQA.ITestCase, error) {

	var test goQA.ITestCase

	test = reflect.New(r.Registry[testName]).Interface().(goQA.ITestCase)	
	test.Init(testName, tm, params)

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
	tm.GetLogger().SetDebug(true)
	tm.AddLogger("console", logger.LOGLEVEL_ALL, console)

	reg := Register{ Registry: regTests}
	tm.RunFromXML("examples\\ExampleTestPlan.xml", &reg)

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\ntotal runtime  = %.6f\n\n", totalTime)
	//time.Sleep(time.Millisecond * 100)
}
