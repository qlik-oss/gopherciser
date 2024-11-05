package session

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
)

const XrfKeyState = "xrfkey-state"

func (state *State) GetXrfKey(host string) (string, error) {
	hostUrl, err := url.Parse(host)
	if err != nil {

		return "", errors.Wrapf(err, "failed to parse host from<%s>", host)
	}

	stateKey := fmt.Sprintf("%s-%s", XrfKeyState, host)

	val, exists := state.GetCustomState(stateKey)
	if exists {
		xrfkey, ok := val.(string)
		if ok {
			return xrfkey, nil
		}
		state.LogEntry.Logf(logger.WarningLevel, "failed to convert xrfkey value<%v> to string, regenerating", val)
	}

	// We don't have one yet, create and add to state and headers
	xrfkey := helpers.GenerateXrfKey(state.Randomizer())
	state.AddCustomState(stateKey, xrfkey)

	headers := state.HeaderJar.GetHeader(hostUrl.Host)
	headers.Add("X-Qlik-XrfKey", xrfkey)

	return xrfkey, nil
}
