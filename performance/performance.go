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
	LE02MillSECOND
	LE03MillSECOND
	LE04MillSECOND
	LE05MillSECOND
	LE10MillSECOND
	LE20MillSECOND
	LE30MillSECOND
	LE01SECOND
	LE02SECOND
	LE03SECOND
	LE04SECOND
	LE05SECOND
	LE10SECOND
	LE20SECOND
	LE30SECOND
	LE01MINUTE
	GT01MINUTE
)

var _RangeName = []string{
	"LE  1  ms",
	"LE  2  ms",
	"LE  3  ms",
	"LE  4  ms",
	"LE  5  ms",
	"LE 10  ms",
	"LE 20  ms",
	"LE 30  ms",
	"LE  1 Sec",
	"LE  2 Sec",
	"LE  3 Sec",
	"LE  4 Sec",
	"LE  5 Sec",
	"LE 10 Sec",
	"LE 20 Sec",
	"LE 30 Sec",
	"LE  1 Min",
	"GT  1 Min",
}

const (
	SECOND30     = 30 * time.Second
	SECOND20     = 20 * time.Second
	SECOND10     = 10 * time.Second
	SECOND05     = 5 * time.Second
	SECOND04     = 4 * time.Second
	SECOND03     = 3 * time.Second
	SECOND02     = 2 * time.Second
	MillSECOND30 = 30 * time.Millisecond
	MillSECOND20 = 20 * time.Millisecond
	MillSECOND10 = 10 * time.Millisecond
	MillSECOND05 = 5 * time.Millisecond
	MillSECOND04 = 4 * time.Millisecond
	MillSECOND03 = 3 * time.Millisecond
	MillSECOND02 = 2 * time.Millisecond
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
	perfData := make([]Statistics, GT01MINUTE+1)
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
		Response:    StatisticsTime{
			MinTime: math.MaxInt64,
		},
		BuildClient: StatisticsTime{
			MinTime: math.MaxInt64,
		},
	}
	outStat.WriteString(fmt.Sprintf("\nperformance statistics :\n"))

	for idx := GT01MINUTE; idx >= 0; idx-- {
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
	if st.Transaction.MinTime  == math.MaxInt64 {
		st.Transaction.MinTime = 0
	}
	if st.Response.MinTime  == math.MaxInt64 {
		st.Response.MinTime = 0
	}
	if st.BuildClient.MinTime  == math.MaxInt64 {
		st.BuildClient.MinTime = 0
	}
	outTimes.WriteString(fmt.Sprintf("-----------------------\nTotal Count: %v\n\n", st.Count))
	outTimes.WriteString(fmt.Sprintf("Use times: %v\n", useTime))
	outTimes.WriteString(fmt.Sprintf("Tps:       %0.3f per/sec\n", float64(time.Second)/float64(useTime)*float64(st.Count)))
	Availability := decimal.NewFromInt(st.Success).Div(decimal.NewFromInt(st.Count)).Mul(decimal.NewFromInt(100))
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
		if respTime < time.Millisecond {
			timeRange = LE01MillSECOND
		} else if respTime < MillSECOND02 {
			timeRange = LE02MillSECOND
		} else if respTime < MillSECOND03 {
			timeRange = LE03MillSECOND
		} else if respTime < MillSECOND04 {
			timeRange = LE04MillSECOND
		} else if respTime < MillSECOND05 {
			timeRange = LE05MillSECOND
		} else if respTime < MillSECOND10 {
			timeRange = LE10MillSECOND
		} else if respTime < MillSECOND20 {
			timeRange = LE20MillSECOND
		} else if respTime < MillSECOND30 {
			timeRange = LE30MillSECOND
		} else if respTime < time.Second {
			timeRange = LE01SECOND
		} else if respTime < SECOND02 {
			timeRange = LE02SECOND
		} else if respTime < SECOND03 {
			timeRange = LE03SECOND
		} else if respTime < SECOND04 {
			timeRange = LE04SECOND
		} else if respTime < SECOND05 {
			timeRange = LE05SECOND
		} else if respTime < SECOND10 {
			timeRange = LE10SECOND
		} else if respTime < SECOND20 {
			timeRange = LE20SECOND
		} else if respTime < SECOND30 {
			timeRange = LE30SECOND
		} else if respTime < time.Minute {
			timeRange = LE01MINUTE
		} else {
			timeRange = GT01MINUTE
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
