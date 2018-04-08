package browser

import (
	"fmt"
	"strings"
)

func IsElementNotFoundErr(err error) bool {
	if err != nil {
		if strings.Contains(err.Error(), `Could not find node with given id`) {
			return true
		}
	}

	return false
}

func TooManyMatchesErr(selector Selector, want int, have int) error {
	return fmt.Errorf(
		"Too many elements matched %q; expected %d, got %d. Please provide a more specific selector.",
		selector,
		want,
		have,
	)
}
