package performance

import "time"

type Tps struct {
	ID         int64
	ReqData    string
	RespStatus int
	STime      time.Time
	SendTime   time.Time
	ETime      time.Time
}

func (tps *Tps) GetID() int64 {
	return tps.ID
}

func (tps *Tps) GetUseTime() (transaction time.Duration, response time.Duration, buildCli time.Duration) {
	transaction = tps.ETime.Sub(tps.STime)
	response = tps.ETime.Sub(tps.SendTime)
	buildCli = tps.SendTime.Sub(tps.STime)
	return
}

func (tps *Tps) GetStatus() int {
	return tps.RespStatus
}

func (tps *Tps) GetData() string {
	return tps.ReqData
}
