package performance

import (
	"bytes"
	"fmt"
	"github.com/shopspring/decimal"
	"go.uber.org/atomic"
	"math"
	"sync"
	"time"
)

const (
	LE01MillSECOND = iota
	LE05MillSECOND
	LE10MillSECOND
	LE100MillSECOND
	LE500MillSECOND
	LE01SECOND
	LE10SECOND
	LE30SECOND
	LE01MINUTE
	LE05MINUTE
	LE10MINUTE
	GT10MINUTE
)

var _RangeName = []string{
	"LE  1  ms",
	"LE  5  ms",
	"LE 10  ms",
	"LE100  ms",
	"LE500  ms",
	"LE  1 Sec",
	"LE 10 Sec",
	"LE 30 Sec",
	"LE  1 Min",
	"LE  5 Min",
	"LE 10 Min",
	"GT 10 Min",
}

const (
	MINUTE10      = 10 * time.Minute
	MINUTE05      = 5 * time.Minute
	MINUTE01      = time.Minute
	SECOND30      = 30 * time.Second
	SECOND01      = time.Second
	MillSECOND500 = 500 * time.Millisecond
	MillSECOND100 = 100 * time.Millisecond
	MillSECOND10  = 10 * time.Millisecond
	MillSECOND05  = 5 * time.Millisecond
	MillSECOND01  = time.Millisecond
)

var lock atomic.Bool
var wg sync.WaitGroup
var _perf *performance

type perfI interface {
	GetID() int64
	GetUseTime() (transaction time.Duration, response time.Duration, buildCli time.Duration)
	GetStatus() int
	GetData() string
}

type performance struct {
	perfChan  chan perfI
	perfData  []Statistics
	dataLock  sync.Mutex
	queueLock sync.Mutex
}

func NewPerformance(routineCnt int64) {
	perfData := make([]Statistics, GT10MINUTE+1)
	_perf = &performance{
		perfChan:  make(chan perfI, routineCnt+1),
		perfData:  perfData,
		dataLock:  sync.Mutex{},
		queueLock: sync.Mutex{},
	}
	go _perf.Run()
	return
}

func Push(perfData perfI) {
	wg.Add(1)
	go func() { _perf.perfChan <- perfData }()
}

func Wait(sTime, eTime time.Time) {
	if lock.Load() {
		return
	}
	lock.Store(true)
	wg.Wait()

	outHead := bytes.Buffer{}
	outStat := bytes.Buffer{}
	outTimes := bytes.Buffer{}
	useTime := eTime.Sub(sTime)
	outHead.WriteString(fmt.Sprintf("Finished at: %v\n", eTime.Format("2006/01/02 15:04:05.999999")))

	cnt := int64(0)
	st := Statistics{
		Transaction: StatisticsTime{
			MinTime: math.MaxInt64,
		},
		Response: StatisticsTime{
			MinTime: math.MaxInt64,
		},
		BuildClient: StatisticsTime{
			MinTime: math.MaxInt64,
		},
	}
	outStat.WriteString(fmt.Sprintf("\nperformance statistics :\n"))

	for idx := GT10MINUTE; idx >= 0; idx-- {
		data := &_perf.perfData[idx]
		if data.Count <= 0 {
			continue
		}
		outStat.WriteString(fmt.Sprintf("%s: count: %d\n", _RangeName[idx], data.Count))
		st.Count += data.Count
		st.Success += data.Success
		st.Failed += data.Failed
		st.Transaction.AddStatisticsTime(&data.Transaction)
		st.Response.AddStatisticsTime(&data.Response)
		st.BuildClient.AddStatisticsTime(&data.BuildClient)

		cnt++
	}
	if st.Transaction.MinTime == math.MaxInt64 {
		st.Transaction.MinTime = 0
	}
	if st.Response.MinTime == math.MaxInt64 {
		st.Response.MinTime = 0
	}
	if st.BuildClient.MinTime == math.MaxInt64 {
		st.BuildClient.MinTime = 0
	}
	tps := float64(0)
	Availability := decimal.Zero
	if st.Count > 0 {
		tps = float64(time.Second) / float64(useTime) * float64(st.Count)
		Availability = decimal.NewFromInt(st.Success).Div(decimal.NewFromInt(st.Count)).Mul(decimal.NewFromInt(100))
	}
	outTimes.WriteString(fmt.Sprintf("-----------------------\nTotal Count: %v\n\n", st.Count))
	outTimes.WriteString(fmt.Sprintf("Use times: %v\n", useTime))
	outTimes.WriteString(fmt.Sprintf("Tps:       %0.3f per/sec\n", tps))
	outTimes.WriteString(fmt.Sprintf("Availability:  %v%%\n", Availability.StringFixed(2)))
	outTimes.WriteString(fmt.Sprintf("Failed:  %v\n", st.Failed))
	outTimes.WriteString(fmt.Sprintf("Connection Times\n                         %-10v     %-10v     %-10v\n", "min", "max", "avg"))
	outTimes.WriteString(fmt.Sprintf("Response time:           %-10v     %-10v     %-10v\n", st.Response.MinTime, st.Response.MaxTime, st.Response.GetAvg()))
	outTimes.WriteString(fmt.Sprintf("Transaction time:        %-10v     %-10v     %-10v\n", st.Transaction.MinTime, st.Transaction.MaxTime, st.Transaction.GetAvg()))
	outTimes.WriteString(fmt.Sprintf("Build cli time:          %-10v     %-10v     %-10v\n", st.BuildClient.MinTime, st.BuildClient.MaxTime, st.BuildClient.GetAvg()))

	fmt.Printf("%v%v%v", outHead.String(), outStat.String(), outTimes.String())
}

func (perf *performance) Run() {
	for perfData := range perf.perfChan {
		useTime, respTime, bTime := perfData.GetUseTime()

		var timeRange int
		if respTime < MillSECOND01 {
			timeRange = LE01MillSECOND
		} else if respTime < MillSECOND05 {
			timeRange = LE05MillSECOND
		} else if respTime < MillSECOND10 {
			timeRange = LE10MillSECOND
		} else if respTime < MillSECOND100 {
			timeRange = LE100MillSECOND
		} else if respTime < MillSECOND500 {
			timeRange = LE500MillSECOND
		} else if respTime < SECOND01 {
			timeRange = LE01SECOND
		} else if respTime < SECOND30 {
			timeRange = LE30SECOND
		} else if respTime < MINUTE01 {
			timeRange = LE01MINUTE
		} else if respTime < MINUTE05 {
			timeRange = LE01MINUTE
		} else if respTime < MINUTE10 {
			timeRange = LE10MINUTE
		} else {
			timeRange = GT10MINUTE
		}

		dataStatus := perfData.GetStatus()
		perf.dataLock.Lock()
		st := &perf.perfData[timeRange]
		st.Count += 1
		st.Transaction.SetTime(useTime)
		st.Response.SetTime(respTime)
		st.BuildClient.SetTime(bTime)
		if dataStatus >= 400 {
			st.Failed += 1
		} else {
			st.Success += 1
		}
		perf.dataLock.Unlock()

		wg.Done()
	}
}
