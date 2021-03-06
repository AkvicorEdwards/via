package permission

import (
	"fmt"
	"testing"
)

func TestParseString(t *testing.T) {
	fmt.Println(ParseString("rw--rw--rw"))
}
