package main

import (
	"fmt"
	//"error"
	//"log"
	"os"
	//"io"
	"github.com/go-QA/goQA"
	"github.com/go-QA/logger"
	"reflect"
	"time"
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
	v1 := t.InitParam("val1", 0.0).(float64)
	v2 := t.InitParam("val2", 0).(int64)
	v3 := t.InitParam("val3", "").(string)

	t.Verify(v1 == 11.11, "verify val1", "Expected 11.11 but got %f instead", v1)
	t.Verify(v2 == 55, "verify val2", "Expected 55 but got %d instead", v2)
	t.Verify(v3 == "hello there test1", "verify val3", "Expected 'hello there test1' but got '%s' instead", v3)

	return t.ReturnFromRun()
}

type Test2 struct {
	data int
	goQA.TestCase
}

func (t *Test2) Run() (int, error) {
	v1 := t.InitParam("val1", 0.0).(float64)
	v2 := t.InitParam("val2", 0).(int64)
	v3 := t.InitParam("val3", "").(string)

	t.Verify(v1 == 111.111, "verify val1", "Expected 111.111 but got %f instead", v1)
	t.Verify(v2 == 550, "verify val2", "Expected 550 but got %d instead", v2)
	t.Verify(v3 == "hello there test2", "verify val3", "Expected 'hello there test2' but got '%s' instead", v3)

	return t.ReturnFromRun()
}

type Test3 struct {
	data int
	goQA.TestCase
}

func (t *Test3) Run() (int, error) {
	v1 := t.InitParam("val1", 0.0).(float64)
	v2 := t.InitParam("val2", 0).(int64)
	v3 := t.InitParam("val3", "").(string)

	t.Verify(v1 == 1111.1111, "verify val1", "Expected 1111.1111 but got %f instead", v1)
	t.Verify(v2 == 5550, "verify val2", "Expected 5550 but got %d instead", v2)
	t.Verify(v3 == "hello there test3", "verify val3", "Expected 'hello there test3' but got '%s' instead", v3)

	return t.ReturnFromRun()
}

var regTests map[string]reflect.Type = map[string]reflect.Type{"test1": reflect.TypeOf(Test1{}),
	"test2": reflect.TypeOf(Test2{}),
	"test3": reflect.TypeOf(Test3{})}

func main() {

	startTime := time.Now()
	//runtime.GOMAXPROCS(2)

	// Report Writer.
	// Only have a TextReporter now that reports plain text to stdout
	tr := goQA.TextReporter{}

	// create the test manager object. Default logger is stdout
	tm := goQA.CreateTestManager(os.Stdout, &tr,
		goQA.SUITE_SERIAL, // Concurency for suites:
		goQA.TC_ALL)       // Concurrency for test cases per suite

	console, err := os.Create("data/console.log")
	if err != nil {
		panic(err)
	}
	defer console.Close()

	//tm.GetLogger().SetDebug(true)

	tm.AddLogger("console", logger.LOG_LEVEL_ALL, console)

	reg := goQA.DefaultRegister{Registry: regTests}

	tm.RunFromXML("examples\\ExampleTestPlan.xml", &reg)

	endTime := time.Now()
	totalTime := endTime.Sub(startTime).Seconds()
	fmt.Printf("\n\ntotal runtime  = %.6f\n\n", totalTime)
	//time.Sleep(time.Millisecond * 100)
}
