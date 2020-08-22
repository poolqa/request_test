package performance

import (
	"github.com/shopspring/decimal"
	"time"
)

type Statistics struct {
	Count       int64
	Success     int64
	Failed      int64
	Transaction StatisticsTime
	Response    StatisticsTime
	BuildClient StatisticsTime
}

type StatisticsTime struct {
	SumTime decimal.Decimal
	//AvgTime time.Duration
	MaxTime time.Duration
	MinTime time.Duration
	Count   int64
}

func (stat *StatisticsTime) AddStatisticsTime(src *StatisticsTime) {
	if stat.SumTime.IsZero() {
		stat.SumTime = src.SumTime
	} else {
		stat.SumTime = stat.SumTime.Add(src.SumTime)
	}
	stat.Count += src.Count
	if src.MaxTime > stat.MaxTime {
		stat.MaxTime = src.MaxTime
	}
	if src.MinTime < stat.MinTime {
		stat.MinTime = src.MinTime
	}
}

func (stat *StatisticsTime) SetTime(src time.Duration) {
	if stat.Count == 0 {
		stat.SumTime = decimal.NewFromInt(int64(src))
		stat.MaxTime = src
		stat.MinTime = src
	} else {
		stat.SumTime = stat.SumTime.Add(decimal.NewFromInt(int64(src)))
		if src > stat.MaxTime {
			stat.MaxTime = src
		}
		if src < stat.MinTime {
			stat.MinTime = src
		}
	}
	stat.Count += 1
}

func (stat *StatisticsTime) GetAvg() time.Duration {
	if stat.SumTime.IsZero() {
		return 0
	}
	return time.Duration(stat.SumTime.Div(decimal.NewFromInt(stat.Count)).IntPart())
}
