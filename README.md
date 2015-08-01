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
	time.Sleep(time.Second * 1)

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
```


