package int64s

import (
	"bytes"
	"fmt"
)

func Join(int64s []int64, sep string) string {
	if len(int64s) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%d", int64s[0]))
	for _, i := range int64s[1:] {
		buffer.WriteString(sep)
		buffer.WriteString(fmt.Sprintf("%d", i))
	}

	return buffer.String()
}
