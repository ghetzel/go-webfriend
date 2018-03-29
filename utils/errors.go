package utils

func IsTimeoutErr(err error) bool {
	if err != nil {
		switch err.Error() {
		case `context deadline exceeded`, `timeout`:
			return true
		}
	}

	return false
}
