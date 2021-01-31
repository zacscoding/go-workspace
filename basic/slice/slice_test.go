package slice

import (
	"fmt"
	"testing"
)

func Test1(t *testing.T) {
	result := returnSlice()
	fmt.Println("## >> In caller Function")
	fmt.Printf("&result: %p ==> %s\n", &result, string(result))
	for i := 0; i < len(result); i++ {
		fmt.Printf("> item[%d] - %p\n", i, &result[i])
	}
	useSlice(result)
	fmt.Println("## >> In caller Function")
	fmt.Printf("&result: %p ==> %s\n", &result, string(result))
	for i := 0; i < len(result); i++ {
		fmt.Printf("> item[%d] - %p\n", i, &result[i])
	}
}

func useSlice(ret []byte) {
	fmt.Println("## >> In args Function")
	fmt.Printf("&result: %p ==> %s\n", &ret, string(ret))
	for i := 0; i < len(ret); i++ {
		fmt.Printf("> item[%d] - %p\n", i, &ret[i])
	}
	ret[0] = 'a'
}

func returnSlice() []byte {
	ret := []byte("test")
	fmt.Println("## >> In returnSlice() Function")
	fmt.Printf("&ret: %p ==> %s\n", &ret, string(ret))
	for i := 0; i < len(ret); i++ {
		fmt.Printf("> item[%d] - %p\n", i, &ret[i])
	}
	return ret
}
