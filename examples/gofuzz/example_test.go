package gofuzz

import (
	"encoding/json"
	"fmt"
	"github.com/google/gofuzz"
	"log"
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	repeat := 1
	f := fuzz.New().NilChance(0)
	log.Println("Test int")
	for i := 0; i < repeat; i++ {
		var intData int
		f.Fuzz(&intData)
		log.Printf("[#%d] > %d", i+1, intData)
	}

	log.Println("Test string")
	for i := 0; i < repeat; i++ {
		var strData string
		f.Fuzz(&strData)
		log.Printf("[#%d] > %s", i+1, strData)
	}
}

func TestBasic2(t *testing.T) {
	unicodeRanges := fuzz.UnicodeRanges{
		{First: 'a', Last: 'z'},
		{First: '0', Last: '9'},
	}
	f := fuzz.New().Funcs(unicodeRanges.CustomStringFuzzFunc())
	repeat := 50
	for i := 0; i < repeat; i++ {
		var strData string
		f.Fuzz(&strData)
		log.Printf("[#%d] > %s", i+1, strData)
	}
}

type OuterStruct struct {
	IntVal    int    `json:"intVal" fuzz:"intValTag"`
	StringVal string `json:"stringVal" fuzz:"stringValTag"`
	Inner     struct {
		Int32Val     int32   `json:"int32Val"`
		Float32Val   float32 `json:"float32Val"`
		StringPtrVal *string `json:"stringPtrVal"`
	} `json:"inner"`
	Inner2 *InnerStruct `json:"inner2"`
}

type InnerStruct struct {
	Int64Val     int64   `json:"int64Val"`
	StringVal    string  `json:"stringVal"`
	StringPtrVal *string `json:"stringPtrVal"`
}

func (s OuterStruct) ToJson() string {
	bytes, err := json.Marshal(&s)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func TestObject(t *testing.T) {
	repeat := 10
	f := fuzz.New().NilChance(0)

	for i := 0; i < repeat; i++ {
		var obj OuterStruct
		f.Fuzz(&obj)
		fmt.Printf("[#%d] %s\n", i+1, obj.ToJson())
	}
	// Output
	//[#1] {"intVal":-2475997130275677392,"stringVal":"Cɛf碵ɐŻ卽ǽ憾{","inner":{"int32Val":1915279670,"float32Val":0.5451799,"stringPtrVal":"©ƽĶkȝɿ诮熹\u003e勳JĶ磼2`5Ơƫ"}}
	//[#2] {"intVal":-7631960148021625142,"stringVal":"06ǩ妌ʚ!籐","inner":{"int32Val":-1734746991,"float32Val":0.6167029,"stringPtrVal":"@lƔ踕N諮ʚ`"}}
	//[#3] {"intVal":-779925702586960001,"stringVal":"蒵ǁ瘧鴃燞-徘繢7Mƺ摁櫂Uj襽","inner":{"int32Val":-1455898371,"float32Val":0.6677701,"stringPtrVal":"杵ąȥ*粠鬦銵ȃ檍Ŝ,福蠚×Ǚ"}}
	//[#4] {"intVal":-7381333340684613686,"stringVal":"l衭Ƌ绛e稸假ý貲v","inner":{"int32Val":-1515145351,"float32Val":0.11166734,"stringPtrVal":"ƯƢŃǒ醆蘐"}}
	//[#5] {"intVal":3788247468831162663,"stringVal":"蕑箺ȱŭ7鰏ȇ裙麅雬杲W\u003e韞ǅ","inner":{"int32Val":1623856972,"float32Val":0.51974,"stringPtrVal":"帗ǀ娑O·rť"}}
	//[#6] {"intVal":-2595571659405522128,"stringVal":"Xd旽[Ⱥ霔(Ʀ雈桯蜴(Ɲ","inner":{"int32Val":-237699371,"float32Val":0.36709365,"stringPtrVal":"(Ǆ·噬CƤQJ½蟿1\u0026瘥"}}
	//[#7] {"intVal":-2538405017249538206,"stringVal":"","inner":{"int32Val":-1804759733,"float32Val":0.8913801,"stringPtrVal":"ǖ/ø垡ɬu蘂MǩȋʧƋlʈ|刏鲧\u003c6"}}
	//[#8] {"intVal":-2245020957079835744,"stringVal":"ǝ","inner":{"int32Val":845393020,"float32Val":0.8788053,"stringPtrVal":"Rw¤ǻăǷ瓡孕Ę4ħ"}}
	//[#9] {"intVal":8632213443791697913,"stringVal":"ǫ5珃/竢ȝşɤ[\u0026摱椔żP","inner":{"int32Val":571681784,"float32Val":0.1842164,"stringPtrVal":""}}
	//[#10] {"intVal":4145896844763383928,"stringVal":"A\\ȏ知ƨ弜sc[ʞZʘƱ","inner":{"int32Val":-1108782606,"float32Val":0.8446675,"stringPtrVal":"蕅嵣蘊惙aʨ疅ǽÒŻ{Ɩ?"}}
}

func TestVisitFields(t *testing.T) {
	f := fuzz.New().NilChance(0)
	obj1 := OuterStruct{
		IntVal: 1,
		Inner2: &InnerStruct{},
	}
	obj2 := OuterStruct{
		IntVal: 2,
		Inner2: &InnerStruct{},
	}
	visitAll(f, reflect.ValueOf(&obj1).Elem())

	fmt.Println("obj1:", obj1.ToJson())
	fmt.Println("obj2:", obj2.ToJson())
}

func visitAll(fuzzer *fuzz.Fuzzer, v reflect.Value) {
	vtype := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fmt.Printf("idx:%d, name:%s, type:%s, value:%v, tag:%s, addressable:%v, caninterface():%v\n",
			i, vtype.Field(i).Name, f.Type(), f.Interface(), vtype.Field(i).Tag.Get("fuzz"),
			f.CanAddr(), f.CanInterface())
		switch f.Kind() {
		case reflect.Struct:
			vs := reflect.ValueOf(f.Interface())
			visitAll(fuzzer, vs)
		default:
			if f.CanAddr() {
				fuzzer.Fuzz(f.Addr().Interface())
			}
		}
	}
}
