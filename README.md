#go-QA
Test framework designed to be used as a base for more specialized test frameworks. For instance, GUI engine, performance/stress, or API testing.

Being written in Go. the framework is very good at running suites and test cases with desired levels of concurrency,  which might make it ideal for stress and performance testing.

###Features:
-Control over concurrency running suites and tests
-logger with levels and logs to multiple io.Write
-Test results
-parameter passing
-Test Manager
-Test Suites
-Test Plan Ran From XML file
 
## Quick Start

To download, run the following command:

~~~
go get github.com/go-QA/goQA.git
~~~



built in and is easily extended and modified. A test case is a struct that inherits another struct, TestCase. TestCase is part of interface, iTestCase, that  has methods the test can override: Init(), Setup(), Run(), and of course TearDown(). A test will at least have the Run() method in most cases.    Have a look at:    /examples/example1.go   This will give a good idea of what the framework is all about

