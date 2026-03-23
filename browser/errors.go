package browser

import (
	"errors"
)

var ExitRequested = errors.New(`exit requested`)
var NotImplemented = errors.New(`Not Implemented`)
var NoActivePage = errors.New(`No active page`)

func IsExitRequestedErr(err error) bool {
	return (err == ExitRequested)
}

func IsNotImplementedErr(err error) bool {
	return (err == NotImplemented)
}
