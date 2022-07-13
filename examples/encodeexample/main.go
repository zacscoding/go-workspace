package main

import (
	"fmt"

	"github.com/linkedin/goavro/v2"
	"github.com/vmihailenco/msgpack/v5"
	"go-workspace/examples/encodeexample/pb"
	"google.golang.org/protobuf/proto"
)

func main() {
	// encodeWithMessagePack()
	// encodeWithProtobuf()
	encodeWithAvro()
}

func encodeWithMessagePack() {
	type Person struct {
		UserName       string
		FavoriteNumber int
		Interests      []string
	}

	person := Person{
		UserName:       "Martin",
		FavoriteNumber: 1337,
		Interests:      []string{"daydreaming", "hacking"},
	}

	b, err := msgpack.Marshal(&person)
	if err != nil {
		panic(err)
	}

	Println("Encode with message pack")
	Printf("%x", b)

	var p Person
	if err := msgpack.Unmarshal(b, &p); err != nil {
		panic(err)
	}
	Println("Unmarshal Record:", p)
	// Output
	//Encode with message pack
	//83a8557365724e616d65a64d617274696eae4661766f726974654e756d626572cd0539a9496e7465726573747392ab646179647265616d696e67a76861636b696e67
	//Unmarshal Record: {Martin 1337 [daydreaming hacking]}
}

func encodeWithProtobuf() {
	favorite := int64(1337)
	person := pb.Person{
		UserName:       "Martin",
		FavoriteNumber: &favorite,
		Interests:      []string{"daydreaming", "hacking"},
	}

	b, err := proto.Marshal(&person)
	if err != nil {
		panic(err)
	}
	Println("Encode with protobuf")
	Printf("%x", b)

	var p pb.Person
	if err := proto.Unmarshal(b, &p); err != nil {
		panic(err)
	}
	Println("Unmarshal Record:", p)
	// Output
	//Encode with protobuf
	//0a064d617274696e10b90a1a0b646179647265616d696e671a076861636b696e67
	//Unmarshal Record: {{{} [] [] 0xc000108160} 0 [] Martin 0xc00001a1b8 [daydreaming hacking]}
}

func encodeWithAvro() {
	codec, err := goavro.NewCodec(`
		{
		  "type": "record",
		  "name": "Person",
		  "fields": [
			{"name": "userName", "type":  "string"},
			{"name": "favoriteNumber", "type":  ["null", "long"], "default": null},
			{"name": "interests", "type":  {"type":  "array", "items":  "string"}}
		  ]
		}`)
	if err != nil {
		panic(err)
	}

	textual := []byte(`{"userName": "Martin",  "favoriteNumber":1337,  "interests": ["daydreaming", "hacking"]}`)

	native, _, err := codec.NativeFromTextual(textual)
	if err != nil {
		panic(err)
	}

	Printf("%x", native)
}

func Println(a ...interface{}) {
	fmt.Println(a...)
}

func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	fmt.Println()
}
