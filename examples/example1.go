package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	"github.com/go-QA/goQA"
	"github.com/go-QA/logger"
	//"runtime"
	"time"
)

// TestCount Number is Number of Test cases used per Suite
const TestCount = 5

// ----------  sample test cases

// doubleIt Example method under test
func doubleIt(d int) int {
	return d * 2
}

// Test1 first test case
// Define a struct that has the goQA.TestCase struct inside.
//You can then add any of the Init(), Run(), and Teardown() methods
//defined in the iTestCase interface
type Test1 struct {
	data int
	goQA.TestCase
}

// Init function defined in Test1 so it can be created and initialized at same time (will see below)
func (tc *Test1) Init(name string, parent goQA.Manager, params goQA.Parameters) goQA.Tester {
	tc.TestCase.Init(name, parent, params)
	return tc
}

// Setup has parameter "val" passed into the object as part of paramList,
//  it will use that value for data if defined or it will use default value of 25
func (tc *Test1) Setup() (int, error) {
	tc.data = tc.InitParam("val", 25).(int)
	return goQA.TcPassed, nil
}

// Run will test Double() function in different ways
func (tc *Test1) Run() (int, error) {
	time.Sleep(time.Millisecond * 100)

	// longhand way to pass/fail a test
	if doubleIt(tc.data) == 20 {
		tc.LogPass("doubleIt(%d) == 20", tc.data)
	} else {
		tc.LogFail("doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	}

	// Perform test in critical section.
	// any failure will cause test to fail and finish running
	// regardless of failure threshold
	tc.Critical.Start()
	if doubleIt(tc.data) == 40 {
		tc.LogPass("doubleIt(%d) == 20", tc.data)
	} else {
		tc.LogFail("doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	}
	tc.Critical.End()

	// this will do same as above
	tc.Verify(doubleIt(tc.data) == 20, fmt.Sprintf("doubleIt(%d) == 20", tc.data), "doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	// this will cause test to fail
	tc.Verify(doubleIt(tc.data) == 30, fmt.Sprintf("doubleIt(%d) check with 30", tc.data), "doubleIt was not 30! returned %d instead", doubleIt(tc.data))
	// returns passed, failed, or error message depending on error threshold and results of run
	return tc.ReturnFromRun()
}

// Teardown used if needed to cleanup after the test.
// will always be called
func (tc *Test1) Teardown() (int, error) {
	return goQA.TcPassed, nil
}

// Test2 will only override the Run()
type Test2 struct {
	data int
	goQA.TestCase
}

// Run will perform a few example tests
func (tc *Test2) Run() (int, error) {
	time.Sleep(time.Millisecond * 100)
	tc.data = 10
	doubled := doubleIt(tc.data)
	tc.Verify(doubled == 20, fmt.Sprintf("doubleIt(%d) == 20", tc.data), "doubleIt( %d ) != 20. Actual = %d", tc.data, doubled)
	tc.Verify(doubled == 30, fmt.Sprintf("doubleIt(%d) == 30", tc.data), "doubleIt( %d ) != 30. Actual = %d", tc.data, doubled)
	return tc.ReturnFromRun()
}

// MySuite is example of creating user defined Suite
//
type MySuite struct {
	goQA.DefaultSuite
}

// Setup method for suite
func (mc *MySuite) Setup() (status int, msg string, err error) {
	// Suite setup
	mc.LogMessage("SUITE(%s) Override Setup for no raisin", mc.Name())
	return goQA.SuiteOk, "My Special Message", nil
}

func main() {

	var t2 *Test2
	startTime := time.Now()
	//runtime.GOMAXPROCS(2)

	// Create a parameter list that can be past into a test case.
	// The parameter named "val" is set to 10 and will be used for data variable in Test1
	// failureThreshold used to determine if tc.ReturnFromRun() will return pass or fail. It is exceptable failure threshold in percent
	paramList := goQA.NewParameters()
	paramList.AddParam("one", "this is one", "First defined parameter")
	paramList.AddParam("val", 10, "value used in test")
	paramList.AddParam("failureThreshold", 60, "Set failure Threshold to 10% for all checks in test case")

	// Report Writer.
	// Only have a TextReporter now that reports plain text to stdout
	tr := goQA.TextReporter{}

	// create the test manager object. Default logger is stdout
	tm := goQA.NewManager(os.Stdout, &tr,
		goQA.SuiteSerial, // Concurency for suites:
		//   SUITE_SERIAL    run one suite at a time
		//   SUITE_ALL       lunch all suites at same time
		//   1...n           run max of n number of suites at one time
		goQA.TcAll) // Concurrency for test cases per suite
	//   TC_SERIAL        run one test case at a time
	//   TC_ALL           launch all suites at same time
	//   1...n            run max of n number of tests at a time

	console, err := os.Create("console.log")
	if err != nil {
		panic(err)
	}
	defer console.Close()
	tm.AddLogger("console", logger.LogLevelAll, console)

	// create two default suite objects
	suite1 := goQA.NewSuite("suite1", &tm, goQA.Parameters{})
	suite2 := goQA.NewSuite("suite2", &tm, goQA.Parameters{})

	// Create custome suite object
	suite3 := MySuite{}
	suite3.Init("Suite 3", &tm, goQA.Parameters{})

	for i := 0; i < TestCount; i++ {

		// this line creates the Test1 object and adds to Suite1 with paramList object
		suite1.AddTest(&Test1{}, fmt.Sprintf("test1_1_%d", i), paramList)

		// Create a new Test2 object and add it to Suite
		suite1.AddTest(&Test2{}, fmt.Sprintf("test2_1_%d", i), goQA.Parameters{})

		// create same tests again for Suite2
		t2 = new(Test2)
		suite2.AddTest(&Test1{}, fmt.Sprintf("test1_2_%d", i), goQA.Parameters{})
		suite2.AddTest(t2, fmt.Sprintf("test2_2_%d", i), paramList)

		suite3.AddTest(&Test1{}, fmt.Sprintf("test1_3_%d", i), goQA.Parameters{}).AddTest(&Test2{}, fmt.Sprintf("test2_3_%d", i), goQA.Parameters{})

	}

	// This will run the t2 test on it's own (last t2 created in the loop)
	// t2.RunTest()

	// here we run just suite1
	//suite1.RunSuite()

	// Add the two suite objects to the test manager
	tm.AddSuite(suite1)
	tm.AddSuite(suite2)
	tm.AddSuite(&suite3)

	//tm.RunSuite("suite1")
	//tm.RunSuite("suite2")
	//tm.RunSuite("suite3")

	// This will run all suites added to tm using the concurrency level set during creation.
	tm.RunAll()

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\ntotal runtime  = %.6f\n\n", totalTime)
	//time.Sleep(time.Millisecond * 100)
}
