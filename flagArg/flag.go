package flagArg

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
)

var FlagArgs *FlagArg

func InitFlag(f *FlagArg) {

	flag.Int64Var(&f.RoutineCount, "c", 1, "go routine(`concurrency`) count at same time.")
	flag.Int64Var(&f.RunCount, "n", 0, "`requests` at every concurrency")
	flag.Int64Var(&f.RunTimeSec, "t", 0, "`timeLimit` Seconds to max. to spend on benchmarking, timer is start at all routine wake up.")

	flag.StringVar(&f.Method, "m", "GET", "http `method` name.")
	flag.StringVar(&f.Url, "u", "", "target `url`")

	flag.BoolVar(&f.DebugLog, "d", false, "show debug log")
	flag.BoolVar(&f.ReportDetail, "r", false, "show those routine's requests detail at report log,\nbut it wall use more and more memory.")
	flag.BoolVar(&f.PreHeat, "w", false, "waiting all routines stand-by.")
	flag.Int64Var(&f.PrintInterval, "i", 1, "print interval for process request count time.")
	flag.BoolVar(&f.Help, "h", false, "this help")

	flag.Usage = f.Usage
}

type FlagArg struct {
	RoutineCount  int64
	RunCount      int64
	RunTimeSec    int64
	Method        string
	Url           string
	TempFile      string
	PreHeat       bool
	DebugLog      bool
	ReportDetail  bool
	PrintInterval int64
	Help          bool
}

func (f *FlagArg) Usage() {
	_, _ = fmt.Fprintf(os.Stderr, `Transactions Request Pressure Test/ v0.0.1
Usage: pressure -c concurrency -n requests -t timeLimit -m method -u url -[dh]

Options:
`)
	flag.PrintDefaults()
}

func (f *FlagArg) CheckFlag() bool {
	if f.Url == "" {
		fmt.Printf("You need to set url.\n")
		return false
	}
	if f.RoutineCount < 1 {
		fmt.Printf("You need to set concurrency >= 1.\n")
		return false
	}

	if f.RunCount == 0 && f.RunTimeSec == 0 {
		fmt.Printf("You need to set request count or time limit.\n")
		return false
	}
	method := strings.ToUpper(f.Method)
	switch method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodPatch:
	case http.MethodDelete:
	case http.MethodConnect:
	case http.MethodOptions:
	case http.MethodTrace:
	default:
		fmt.Printf("the method [%v] is wrong\n", f.Method)
		return false
	}
	if f.RunTimeSec > 0 {
		f.RunCount = math.MaxInt64
	}
	if f.PrintInterval < 1 {
		f.PrintInterval = 1
	}
	return true
}


