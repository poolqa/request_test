package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math/rand"
	"os"
	"os/signal"
	"request_test/bot"
	"request_test/flagArg"
	"request_test/performance"
	"sync"
	"time"
)

var (
	BotArray  []BotInterface
	c         chan os.Signal
	startTime time.Time
	endTime   time.Time
)

type BotInterface interface {
	StandBy(context.Context, *sync.WaitGroup, int64)
	GetRunCount() (int64, int64)
}

func init() {
	flagArg.FlagArgs = &flagArg.FlagArg{}
	flagArg.InitFlag(flagArg.FlagArgs)
}

func closeBySignal() {
	<-c
	endTime = time.Now()
	performance.Wait(startTime, endTime)
	os.Exit(0)
}

func main() {
	flag.Parse()
	if flagArg.FlagArgs.Help {
		flag.Usage()
		os.Exit(0)
	}
	if !flagArg.FlagArgs.CheckFlag() {
		flag.Usage()
		os.Exit(0)
	}
	rand.Seed(time.Now().Unix())
	c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go closeBySignal()

	fmt.Printf("\nThread count: %v\n", flagArg.FlagArgs.RoutineCount)

	performance.NewPerformance(flagArg.FlagArgs.RoutineCount)

	BotArray = make([]BotInterface, flagArg.FlagArgs.RoutineCount)

	httpClient := &fasthttp.Client{
		MaxConnsPerHost: int(flagArg.FlagArgs.RoutineCount),
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	if flagArg.FlagArgs.ProxyAddr != "" {
		httpClient.Dial = fasthttpproxy.FasthttpHTTPDialer(flagArg.FlagArgs.ProxyAddr)
	}
	wg := &sync.WaitGroup{}
	done := make(chan struct{})
	ctx, runRoutine := context.WithCancel(context.Background())
	if !flagArg.FlagArgs.PreHeat {
		startTime = time.Now()
		fmt.Printf("Starting at: %v\n", startTime.Format("2006/01/02 15:04:05.999999"))
		runRoutine()
	}

	// request routine start
	for idx := int64(0); idx < flagArg.FlagArgs.RoutineCount; idx++ {
		BotArray[idx] = &bot.OrderBot{
			ID:            idx + 1,
			Client:        httpClient,
			SleepInterval: time.Duration(flagArg.FlagArgs.SleepInterval) * time.Millisecond,
		}
		wg.Add(1)
		go BotArray[idx].StandBy(ctx, wg, flagArg.FlagArgs.RunCount)
	}
	// request routine end

	if flagArg.FlagArgs.PreHeat {
		startTime = time.Now()
		fmt.Printf("Starting at: %v\n", startTime.Format("2006/01/02 15:04:05.999999"))
		runRoutine()
	}
	go func() {
		wg.Wait()
		done <- struct{}{}
	}()
	var sendCnt int64
	var doneCnt int64
	var tmpReqCnt int64
	var tmpRespCnt int64
	var stop <-chan time.Time
	ticker := time.NewTicker(time.Duration(flagArg.FlagArgs.PrintInterval) * time.Second)
	if flagArg.FlagArgs.RunTimeSec > 0 {
		stop = time.After(time.Duration(flagArg.FlagArgs.RunTimeSec) * time.Second)
	} else {
		stop = make(<-chan time.Time) // never do it
	}
	var sec = flagArg.FlagArgs.PrintInterval
loop:
	for {
		select {
		case <-done:
			break loop
		case <-stop:
			break loop
		case <-ticker.C:
			sendCnt = 0
			doneCnt = 0
			for _, botI := range BotArray {
				tmpReqCnt, tmpRespCnt = botI.GetRunCount()
				sendCnt += tmpReqCnt
				doneCnt += tmpRespCnt
			}
			p := message.NewPrinter(language.English)
			p.Printf("sec %v, send request count: %d, done: %d\n", sec, sendCnt, doneCnt)
			sec += flagArg.FlagArgs.PrintInterval
		}
	}
	endTime = time.Now()

	performance.Wait(startTime, endTime)
}
