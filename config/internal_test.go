package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/qlik-oss/gopherciser/globals"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/statistics"
)

func TestConfigSummary(t *testing.T) {
	log := logger.NewLog(logger.LogSettings{
		Traffic: false,
		Metrics: false,
		Debug:   false,
	})
	log.AddLoggers(logger.CreateStdoutLogger())

	startTime := time.Now().Add(-5 * time.Minute)

	// clean summaries

	fmt.Println("simple (clean):")
	summary(log, SummaryTypeSimple, startTime)
	fmt.Println()

	fmt.Println("extended (clean):")
	summary(log, SummaryTypeExtended, startTime)
	fmt.Println()

	fmt.Println("full (clean):")
	summary(log, SummaryTypeFull, startTime)
	fmt.Println()

	// "dirty" summaries

	// make dirty
	statistics.SetGlobalLevel(statistics.StatsLevelOn)
	globals.Errors.Add(3)
	globals.Warnings.Add(23)
	globals.Users.Add(6)
	globals.Threads.Add(4)
	globals.Sessions.Add(666)
	statistics.IncOpenedApps()
	statistics.IncCreatedApps()

	openStats := statistics.GetOrAddGlobalActionStats("openapp", "open my cool app", "dc2c9170-871d-4093-b4cf-df6c1fcb1c01")
	openStats.RespAvg.AddSample(uint64(time.Millisecond * 10))
	openStats.Received.Add(123)
	openStats.Sent.Add(123)
	openStats.ErrCount.Add(1)
	openStats.WarnCount.Add(1)
	openStats.Requests.Add(123)

	chStats := statistics.GetOrAddGlobalActionStats("changesheet", "very very very very very very very very very very very very very long label", "dc2c9170-871d-4093-b4cf-df6c1fcb1c01")
	chStats.RespAvg.AddSample(uint64(time.Minute*5 + 4*time.Millisecond))
	chStats.Received.Add(123)
	chStats.Sent.Add(123)
	chStats.ErrCount.Add(1)
	chStats.WarnCount.Add(1)
	chStats.Requests.Add(123)

	fmt.Println("simple (dirty):")
	summary(log, SummaryTypeSimple, startTime)
	fmt.Println()

	fmt.Println("extended (dirty):")
	summary(log, SummaryTypeExtended, startTime)
	fmt.Println()

	statistics.SetGlobalLevel(statistics.StatsLevelFull)
	usersStats := statistics.GetOrAddGlobalRequestStats("GET", "/api/v1/user/me")
	usersStats.RespAvg.AddSample(uint64(time.Millisecond * 400))
	usersStats.Received.Add(2)
	usersStats.Sent.Add(432)

	fmt.Println("full (dirty):")
	summary(log, SummaryTypeFull, startTime)
	fmt.Println()

	// Reset global counter to not effect other tests
	globals.Errors.Reset()
	globals.Warnings.Reset()
	globals.Users.Reset()
	globals.Threads.Reset()
	globals.Sessions.Reset()
	statistics.DestroyGlobalCollector()
}
