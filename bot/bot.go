package bot

import (
	"context"
	"fmt"
	"github.com/valyala/fasthttp"
	"request_test/flagArg"
	"request_test/performance"
	"runtime"
	"strings"
	"sync"
	"time"
)

type OrderBot struct {
	ID       int64
	Client   *fasthttp.Client
	procCnt  int64
	gId      string
}

func (bot *OrderBot) GetRoutineId() string {
	var (
		buf [64]byte
		n   = runtime.Stack(buf[:], false)
		stk = strings.TrimPrefix(string(buf[:n]), "goroutine ")
	)

	idField := strings.Fields(stk)[0]
	return idField
}

func (bot *OrderBot) GetRunCount() int64 {
	return bot.procCnt
}

func (bot *OrderBot) StandBy(ctx context.Context, wg *sync.WaitGroup, runCnt int64) {
	bot.gId = bot.GetRoutineId()
	if flagArg.FlagArgs.DebugLog {
		fmt.Printf("routine:%v start %v...\n", bot.gId, time.Now().Format("2006/01/02 15:04:05.999999"))
	}
	<- ctx.Done()

	for bot.procCnt = int64(1); bot.procCnt <= runCnt; bot.procCnt++ {
		tps := &performance.Tps{
			ID: bot.ID,
			STime: time.Now(),
		}
		// send msg
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()

		req.Header.SetMethod(flagArg.FlagArgs.Method)
		req.SetRequestURI(flagArg.FlagArgs.Url)

		tps.SendTime = time.Now()
		err := bot.Client.Do(req, resp)
		tps.ETime = time.Now()
		if err != nil {
			tps.RespStatus = fasthttp.StatusBadRequest
		} else {
			tps.RespStatus = resp.StatusCode()
		}
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)

		performance.Push(tps)
	}
	wg.Done()
}
