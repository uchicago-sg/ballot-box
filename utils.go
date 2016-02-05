package voting

import (
	"strings"
)

func RouteURL(path string) (string, string, string) {
	parts := strings.SplitN(path[1:], ".", 2)
	ext := ""
	verb := ""

	if len(parts) == 2 {
		ext = parts[1]
	}

	parts = strings.SplitN(parts[0], "/", 2)

	if len(parts) == 2 {
		verb = parts[1]
	}

	return parts[0], verb, ext
}
