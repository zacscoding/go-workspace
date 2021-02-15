package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	ret := []string{
		"sarama-4e2db5b6-9193-4780-83e6-0f7bf5890ab2",
		"sarama-67373826-626a-4976-906c-d1698bdd4665",
	}
	sort.Slice(ret, func(i, j int) bool {
		return strings.Compare(ret[i], ret[j]) < 0
	})
	fmt.Println(strings.Join(ret, ","))
}
