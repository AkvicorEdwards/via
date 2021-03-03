package record

import (
	"fmt"
	"time"
)

func TimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func AddTimeNow(s string) string {
	return fmt.Sprintf("%s %s", TimeNow(), s)
}
