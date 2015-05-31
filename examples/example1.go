package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	//"gitorious.org/goqa/goqa.git"
	"../../goQA"
	//"runtime"
	"time"
)

const TEST_COUNT = 2

// ----------  sample test cases

// Example method under test
func doubleIt(d int) int {
	return d * 2
}

// Defining first test case
// Define a struct that has the goQA.TestCase struct inside.
//You can then add any of the Init(), Run(), and Teardown() methods
//defined in the iTestCase interface
type Test1 struct {
	data int
	goQA.TestCase
}

// Init function defined in Test1 so it can be created and initialized at same time (will see below)
func (tc *Test1) Init(name string, parent goQA.ITestManager, params goQA.Parameters) goQA.ITestCase {
	tc.TestCase.Init(name, parent, params)
	return tc
}

// if parameter "val" is defined and passed into the object as a paramList,
//  it will use that value for data; else it will use default value 25
func (tc *Test1) Setup() (int, error) {
	tc.data = tc.InitParam("val", 25).(int)
	return goQA.TC_PASSED, nil
}

func (tc *Test1) Run() (int, error) {
	time.Sleep(time.Second * 1)

	// longhand way to pass/fail a test
	if doubleIt(tc.data) == 20 {
		tc.TestCase.LogPass("doubleIt(%d) == 20", tc.data)
	} else {
		tc.TestCase.LogFail("doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	}

	// Perform test in critical section.
	// any failure will cause test to fail regardless of
	// failure threshold
	tc.TestCase.Critical.Start()
	if doubleIt(tc.data) == 40 {
		tc.TestCase.LogPass("doubleIt(%d) == 20", tc.data)
	} else {
		tc.TestCase.LogFail("doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	}
	tc.TestCase.Critical.End()

	// this will do same as above
	tc.Verify(doubleIt(tc.data) == 20, fmt.Sprintf("doubleIt(%d) == 20", tc.data), "doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	// this will cause test to fail
	tc.Verify(doubleIt(tc.data) == 30, fmt.Sprintf("doubleIt(%d) check with 30", tc.data), "doubleIt was not 30! returned %d instead", doubleIt(tc.data))
	// returns passed, failed, or error message depending on error threshold and results of run
	return tc.ReturnFromRun()
}

func (tc *Test1) Teardown() (int, error) {
	return goQA.TC_PASSED, nil
}

// Test2 will only override the Run()
type Test2 struct {
	data int
	goQA.TestCase
}

func (tc *Test2) Run() (int, error) {
	time.Sleep(time.Second * 1)
	tc.data = 10
	doubled := doubleIt(tc.data)
	tc.Verify(doubled == 20, fmt.Sprintf("doubleIt(%d) == 20", tc.data), "doubleIt( %d ) != 20. Actual = %d", tc.data, doubled)
	tc.Verify(doubled == 30, fmt.Sprintf("doubleIt(%d) == 30", tc.data), "doubleIt( %d ) != 30. Actual = %d", tc.data, doubled)
	return tc.ReturnFromRun()
}

// new struct to create Suite Contains DefaultSuite struct
type MySuite struct {
	goQA.DefaultSuite
}

// Define our own Setup method
func (mc *MySuite) Setup() (status int, msg string, err error) {
	// Suite setup
	fMsg := fmt.Sprintf("SUITE(%s) Override Setup for no raisin", mc.Name())
	mc.TestCase.LogMessage(fMsg)
	//g.logger.Printf("PASS::%s\n", passMsg)
	return goQA.SUITE_OK, "My Special Message", nil
}

func main() {

	var t2 *Test2
	startTime := time.Now()
	//runtime.GOMAXPROCS(2)

	// Create a parameter list that can be past into a test case.
	// The parameter named "val" is set to 10 and will be used for data variable in Test1
	// failureThreshold used to determine if tc.ReturnFromRun() will return pass or fail. It is exceptable failure threshold in percent
	paramList := goQA.CreateParametersObject()
	paramList.AddParam("one", "this is one", "First defined parameter")
	paramList.AddParam("val", 10, "value used in test")
	paramList.AddParam("failureThreshold", 60, "Set failure Threshold to 10% for all checks in test case")

	// Report Writer.
	// Only have a TextReporter now that reports plain text to stdout
	tr := goQA.TextReporter{}

	// create the test manager object. Default logger is stdout
	tm := goQA.CreateTestManager(os.Stdout, &tr,
		goQA.SUITE_SERIAL, // Concurency for suites:
		//   SUITE_SERIAL    run one suite at a time
		//   SUITE_ALL       lunch all suites at same time
		//   1...n           run max of n number of suites at one time
		goQA.TC_ALL) // Concurrency for test cases per suite
	//   TC_SERIAL        run one test case at a time
	//   TC_ALL           launch all suites at same time
	//   1...n            run max of n number of tests at a time

	console, err := os.Create("data/console.log")
	if err != nil {
		panic(err)
	}
	defer console.Close()
	tm.AddLogger("console", goQA.LOGLEVEL_ALL, console)

	// create two suite objects
	suite1 := goQA.CreateSuite("suite1", &tm, goQA.Parameters{})
	// User defined suite
	suite2 := MySuite{}
	suite2.Init("suite2", &tm, goQA.Parameters{})

	for i := 0; i < TEST_COUNT; i++ {
		// create Test2 object and call Init() the Init must be called before running test
		t2 = new(Test2)
		t2.Init(fmt.Sprintf("test2_1_%d", i), &tm, goQA.Parameters{})
		// add test2 object to suite1
		suite1.AddTest(t2)

		// this line creates the Test1 object, calls the Test1.Init() method, then adds test to suite1. Notice the paranList object is passed to test
		suite1.AddTest(new(Test1).Init(fmt.Sprintf("test1_1_%d", i), &tm, paramList))

		// create same tests again for suite 2 but pass the paramList object to Test2 instead (We're having fun now right?)
		t2 = new(Test2)
		t2.Init(fmt.Sprintf("test2_2_%d", i), &tm, paramList)
		suite2.AddTest(t2)
		suite2.AddTest(new(Test1).Init(fmt.Sprintf("test1_2_%d", i), &tm, goQA.Parameters{}))
	}

	// This will run the t2 test on it's own (last t2 created in the loop)
	t2.RunTest(t2)

	// here we run just suite1
	suite1.RunSuite()

	// Add the two suite objects to the test manager
	tm.AddSuite(suite1)
	tm.AddSuite(&suite2)

	// This will run all suites added to tm using the concurrency level set during creation.
	tm.RunAll()
	//tm.RunFromXML("ChamberFunctionality.xml")
	//console.Sync()
	//console.Close()

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\ntotal runtime  = %.6f\n\n", totalTime)
	//time.Sleep(time.Millisecond * 100)
}