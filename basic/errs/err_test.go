package errs

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"testing"
)

func TestWrap(t *testing.T) {
	depth := 3
	err := ErrorWraps(depth)
	cause := errors.Cause(err)
	fmt.Println("err ::", err)
	fmt.Println("cause ::", cause)
	// Output
	//err :: depth-1: depth-2: depth-3: this is cause!
	//cause :: this is cause!
}


func ErrorWraps(depth int) error {
	cause := errors.New("this is cause!")
	err := cause

	for i := depth; i > 0; i-- {
		err = errors.Wrap(err, "depth-"+strconv.Itoa(i))
	}
	return err
}
