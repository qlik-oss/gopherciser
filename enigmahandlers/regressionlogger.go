package enigmahandlers

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/qlik-oss/gopherciser/logger"
)

var myFile = initRegressionLog("logs/oooooooo.regression.log")
var enableRegressionLog = true

type Direction string

const (
	Sent     Direction = "SENT"
	Recieved Direction = "RECIEVED"
)

type Protocol string

const (
	Unknown Protocol = "UNKNOWN"
	WS      Protocol = "WS"
	HTTP    Protocol = "HTTP"
)

const templateStr = `MSG_{{.Msg.Idx}} {{Msg.Protocol}} {{Msg.Direction}}
ACTION id={{.Action.ID}} name="{{Action.Name}}"

{{Msg.Payload}}
MSG_{{.Msg.Idx}} END
`

var msgCnt uint64 = 0

func LogRegression(logEntry *logger.LogEntry, direction Direction, message []byte) {
	if enableRegressionLog {
		msgIdx := atomic.AddUint64(&msgCnt, 1)
		protocol := Unknown
		switch {
		case len(message) > 0 && message[0] == byte('{'):
			protocol = WS
		default:
			protocol = HTTP
		}
		fmt.Fprintf(myFile, "MSG_%d %s %s\n", msgIdx, protocol, direction)
		fmt.Fprintf(myFile, `ACTION id=%d name="%s" label="%s"\n"`, logEntry.Action.ActionID, logEntry.Action.Action, logEntry.Action.Label)
		fmt.Fprintf(myFile, "%s\n", message)
		fmt.Fprintf(myFile, "MSG_%d END\n")
	}
}

func initRegressionLog(fileName string) *os.File {
	myFile, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	return myFile
}
