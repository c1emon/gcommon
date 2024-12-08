package data

import (
	"fmt"
	"strings"
)

func warpDoubleQuotes(raw string) string {
	return fmt.Sprintf("\"%s\"", strings.Trim(raw, `"`))
}
