package voting

import (
	"strconv"
	"strings"
)

func RouteURL(path string) (int64, string, string) {
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

  parts = strings.Split(parts[0], "-")
	eid, _ := strconv.ParseInt(parts[0], 10, 32)

	return eid, verb, ext
}
