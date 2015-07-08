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

####Create Test:


