package browser

import (
	"errors"
)

var ExitRequested = errors.New(`exit requested`)
var NotImplemented = errors.New(`Not Implemented`)

func IsExitRequestedErr(err error) bool {
	return (err == ExitRequested)
}

func IsNotImplementedErr(err error) bool {
	return (err == NotImplemented)
}
