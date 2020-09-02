package main

import (
	"fmt"
	"github.com/speps/go-hashids"
)

// tests go-hashids to en/decoding betwwen numbers and string
func main() {
	hd := hashids.NewData()
	hd.Salt = "this is my salt"

	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}

	for i := 1; i < 10; i++ {
		encoded, err := h.Encode([]int{i})
		if err != nil {
			fmt.Printf("%d --> error:%v\n", i, err)
		}
		decoded, err := h.DecodeWithError(encoded)
		if err != nil {
			fmt.Printf("%d <-- %s error:%v\n", i, encoded, err)
		}

		if i == decoded[0] {
			fmt.Printf("%d <--> %s\n", i, encoded)
		} else {
			fmt.Printf("%d != %s\n", i, encoded)
		}
	}

	// Output
	//1 <--> NV
	//2 <--> 6m
	//3 <--> yD
	//4 <--> 2l
	//5 <--> rD
	//6 <--> lv
	//7 <--> jv
	//8 <--> D1
	//9 <--> Qg
}
