package fetchque

import "fmt"

type Data1 struct {
	RequestedKey string
	Value        string
}

func (d Data1) String() string {
	return fmt.Sprintf("Data1{RequestedKey:%s, Value:%s}", d.RequestedKey, d.Value)
}

type Data2 struct {
	RequestedKey string
	Value        string
}

func (d Data2) String() string {
	return fmt.Sprintf("Data2{RequestedKey:%s, Value:%s}", d.RequestedKey, d.Value)
}
