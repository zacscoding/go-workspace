package gofuzz

import (
	fuzz "github.com/google/gofuzz"
	"reflect"
)

// ComplexFuzzer injects fuzz data
type ComplexFuzzer struct {
	defaultFuzzer *fuzz.Fuzzer
	customFuzzers map[string]*fuzz.Fuzzer
}

func (cf *ComplexFuzzer) Fuzz(obj interface{}) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		panic("needed ptr!")
	}
}
