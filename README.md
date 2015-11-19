#go-QA
Test framework designed to be used as a base for more specialized test frameworks. For instance, GUI engine, performance/stress, or API testing.

Since it is written in Go. the framework is very good at running suites and test cases with desired levels of concurrency,  which might make it ideal for stress and performance testing.

##Features:
- Control over concurrency running suites and tests (serial, throttled, or all at once)
- logger with levels and logs to multiple io.Write
- Test results
- parameter passing
- Test Manager
- Test Suites
- Test Plan Ran From XML file
 
## Quick Start

To download, run the following command:

~~~
go get github.com/go-QA/goQA
go get github.com/go-QA/logger
~~~


##Create and Run a test case in a suite:

  There is a more detailed and running example of this explanation below located at `/examples/example1.go`

 A test case is part of interface `ITestCase` to run from `TestManager`. 

```go
 type iTestCase interface {
	Name() string
	Init(name string, parent ITestManager, params Parameters) ITestCase
	Setup() (int, error)
	Run() (int, error)
	Teardown() (int, error)
}
```

  This example test case will unit test this overly complicated amazing method:
  ```go
 // Example method under test
func doubleIt(d int) int {
	return d * 2
}
```

So, let's define a struct that has the `goQA.TestCase` struct inside.
This struct already is part of 'ITestCase' interface and has logger, test methods, critical section,
parmeter handling, and all the good stuff you need.
YYou need to call the `Init()` before running and should at least override `Run()` so it does
something.

 ###Create a test case struct:

 ```go
type Test1 struct {
	data int
	goQA.TestCase
}
```

First, let's override `Setup()` to initialize the param "val".
The `tc.data` will be initialized to param value passed in or
use the default value of 25 in no parameter for "val" is specified

```go
func (tc *Test1) Setup() (int, error) {
	tc.data = tc.InitParam("val", 25).(int)
	return goQA.TC_PASSED, nil
}
```
Next, create the `Run()` method for the test. 

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


The `goQA.Parameters` is used to pass the test parameters into our test. 
```go
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
	// Creates a new default Suite object. You can define your own suites as well using
    // the goQA.ISuite interface.
	suite1 := goQA.CreateSuite("suite1", &tm, goQA.Parameters{})
```

Now simply create an object of Test1 and call the Init() method. The object is then added to suite1 object created above.

```go
	// create Test1 object and call Init(), The Init must be called before running test
	test := new(Test1)
	test.Init("test 1", &tm, paramList)

	// add trst case to suite
	suite1.AddTest(test)
```

 The suite is added to TestManager :
```go
	// Add the two suite objects to the test manager
	tm.AddSuite(suite1)
```

Here are three examples of ways to launch the test case manually:

```go
	// This will run the test on it's own using the TestManager reference passed in
	test.RunTest(test)

	// here we run suite1 test cases
	suite1.RunSuite()

	// This will run all suites added to tm using the concurrency level set during creation.
	tm.RunAll()
```


##Run From XML Test Plan

  A test plan can be created with an XML file and ran by the `TestManager` by calling `RunFromXML(File, Register)`
There is a working example `goQA/Examples/example_RunFromXML.go`

 This is the example XML file that creates Three test cases for a suite:

```xml
<?xml version="1.0"?>
<TestManager name="Manager">
   <TestSuite name="suite1">
    <Param name='Domain' type='string'>github.com/go-QA/goQA</Param>
     <TestCase name="test1">
      <Param name='MaxTime' type='int' value='300' comment='Set Max Tme to run test'/>
      <Param name='val1' type='float' value='111.111' comment='float value'/>
      <Param name='val2' type='int' value='550' comment="val1 is integer"/>
      <Param name='val2' type='string' value='hello there' comment='val2 is string'/>
    </TestCase>
     <TestCase name="test2">
      <Param name='MaxTime' type='int' value='3040' comment='Set Max Tme to run test'/>
      <Param name='val1' type='float' value='1141.111' comment='float value'/>
      <Param name='val2' type='int' value='5504' comment="val1 is integer"/>
      <Param name='val2' type='string' value='hello there' comment='val2 is string'/>
    </TestCase>
      <TestCase name="test3">
      <Param name='MaxTime' type='int' value='3005' comment='Set Max Tme to run test'/>
      <Param name='val1' type='float' value='1151.111' comment='float value'/>
      <Param name='val2' type='int' value='5505' comment="val1 is integer"/>
      <Param name='val2' type='string' value='hello there' comment='val2 is string'/>
    </TestCase>
  </TestSuite>
</TestManager>
```


 To create a test plan you must have a type that implements the `goQA.Register` interface to create the concrete test cases and suite objects:

```go
	type TestRegister interface {
		GetTestCase(testType string, tm *TestManager, params Parameters) (ITestCase, error)
		GetSuite(testType string, tm *TestManager, params Parameters) (Suite, error)
	}
```

this is from example file; a basic way to register your test cases:

```go
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
```

From here you call the manager object:

```go
	tm.RunFromXML(filePath, &reg)
```

