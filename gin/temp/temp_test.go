package temp

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	cases := []string{
		"", "1", "2", "3", "aa",
	}

	for _, s := range cases {
		testTemp(s)
	}
}

func testTemp(val string) {
	switch val {
	case "":
	case "1":
	case "2":
	default:
		return
	}
	fmt.Println("Execute...:", val)
}
