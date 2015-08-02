#go-QA
Test framework designed to be used as a base for more specialized test frameworks. For instance, GUI engine, performance/stress, or API testing.

Since it is written in Go. the framework is very good at running suites and test cases with desired levels of concurrency,  which might make it ideal for stress and performance testing.

##Features:
- Control over concurrency running suites and tests
- logger with levels and logs to multiple io.Write
- Test results
- parameter passing
- Test Manager
- Test Suites
- Test Plan Ran From XML file
 
### Quick Start

To download, run the following command:

~~~
go get github.com/go-QA/goQA.git
~~~


 A test case is part of interface `ITestCase` to run from `TestManager`

```go
 type iTestCase interface {
	Name() string
	Init(name string, parent ITestManager, params Parameters) ITestCase
	Setup() (int, error)
	Run() (int, error)
	Teardown() (int, error)
}
```


####Create Test:

#Defining first test case:

  This test case will test unit test this overly complicated method:
  ```go
 // Example method under test
func doubleIt(d int) int {
	return d * 2
}
```

Define a struct that has the `goQA.TestCase` struct inside.
This struct already is part of 'ITestCase' interface and has logger, test methods, critical section,
parmeter handling, and all the good stuff you need.
YYou need to call the `Init()` before running and should at least override `Run()` so it does
something.

 Create a test case as a truct:

 ```go
type Test1 struct {
	data int
	goQA.TestCase
}
```

First, let's override `Setup()` to initialize the param "val".
The `tc.data` will be initialized to param value passed in or default value of 25
```go
func (tc *Test1) Setup() (int, error) {
	tc.data = tc.InitParam("val", 25).(int)
	return goQA.TC_PASSED, nil
}
```
Next, create the run method for the test. 

```go
func (tc *Test1) Run() (int, error) {

	// longhand way to pass/fail a test
	if doubleIt(tc.data) == 20 {
		tc.LogPass("doubleIt(%d) == 20", tc.data)
	} else {
		tc.LogFail("doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	}

	// this will do same as above
	tc.Verify(doubleIt(tc.data) == 20, fmt.Sprintf("doubleIt(%d) == 20", tc.data), "doubleIt( %d ) != 20. Actual = %d", tc.data, doubleIt(tc.data))
	// this will cause test to fail
	tc.Verify(doubleIt(tc.data) == 30, fmt.Sprintf("doubleIt(%d) check with 30", tc.data), "doubleIt was not 30! returned %d instead", doubleIt(tc.data))
	// returns passed, failed, or error message depending on error threshold and results of run
	return tc.ReturnFromRun()
}

```

  A test requires a TestManager object to run. The TestManager needs a reporter object passed in we will create. A Suite will also be created for this example.

 Let's get all the boiler plate to create a TestManager object, `tm`, to use with our test:

```go
func main() {

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

```
  The `goQA.TextReporter` is a simple reporter that will display test results for suites and test cases as a simple text output. Need better reporting in the near future. 


	// Create a parameter list that can be past into a test case.
	// - The parameter named "val" is set to 10 and will be used for data variable in Test1
	// - failureThreshold used to determine if tc.ReturnFromRun() will return pass or fail. 
    //     Test will report failed if more than 10% of checks fail during run
	paramList := goQA.CreateParametersObject()
	paramList.AddParam("val", 10, "value used in test")
	paramList.AddParam("failureThreshold", 60, "Set failure Threshold to 10% for all checks in test case")

```

next, a default Suite object is created 

```go
	// create two suite objects
	suite1 := goQA.CreateSuite("suite1", &tm, goQA.Parameters{})
```

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

	console, err := os.Create("console.log")
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


