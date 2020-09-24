package bind

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

// We can extract ["field", "value", "reason"] if provide valid type value
// such as field type is uint and provided value is "1". But if provide value "aa" i.e non numeric,
// we cannot handle error reason because gin.ShouldBind returns strconv
func TestShouldBind(t *testing.T) {
	go StartGinServer()

	cases := []struct {
		Name      string
		Path      string
		QueryHex  string
		QuerySize string
	}{
		////////////////////////////////////////////////
		// bind/should-bind
		{
			Name:      "[bind/should-bind] Required hex but empty",
			Path:      "bind/should-bind",
			QuerySize: "10",
		}, {
			Name:      "[[bind/should-bind]] Required hexadecimal format",
			Path:      "bind/should-bind",
			QueryHex:  "query hex aza",
			QuerySize: "10",
		}, {
			Name:      "[[bind/should-bind]] Size must be greater then or equals to 3",
			Path:      "bind/should-bind",
			QueryHex:  "a",
			QuerySize: "1",
		}, {
			Name:      "[[bind/should-bind]] Size must be numeric",
			Path:      "bind/should-bind",
			QueryHex:  "a",
			QuerySize: "a",
		},
		////////////////////////////////////////////////
		// bind/should-bind2
		{
			Name:      "[bind/should-bind2] Size must be greater then or equals to 3",
			Path:      "bind/should-bind2",
			QuerySize: "3",
		}, {
			Name:      "[bind/should-bind2] Size must be greater then or equals to 3",
			Path:      "bind/should-bind2",
			QueryHex:  "a",
			QuerySize: "1",
		}, {
			Name:      "[bind/should-bind2] Size must be numeric",
			Path:      "bind/should-bind2",
			QueryHex:  "a",
			QuerySize: "a",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			query := make(map[string]string)
			putIfNotEmpty(query, "hex", tc.QueryHex)
			putIfNotEmpty(query, "size", tc.QuerySize)

			u := &url.URL{
				Scheme: "http",
				Host:   "localhost:3000",
				Path:   tc.Path,
			}
			q := u.Query()
			for k, v := range query {
				q.Set(k, v)
			}
			u.RawQuery = q.Encode()

			req, err := http.NewRequest("GET", u.String(), nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			fmt.Println("Request URI:", u.String())

			if err != nil {
				fmt.Println("> Error:", err)
				fmt.Println("----------------------------------------------------------------------------------")
				return
			}
			defer resp.Body.Close()
			b, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("> Body:", string(b))
			fmt.Println("----------------------------------------------------------------------------------")
		})
	}
	// Output
	//Request URI: http://localhost:3000/bind/should-bind?size=10
	//> Body: {"errors":[{"field":"hex","message":"required hex","value":""}]}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Required_hexadecimal_format
	//Request URI: http://localhost:3000/bind/should-bind?hex=query+hex+aza&size=10
	//> Body: {"errors":[{"field":"hex","message":"required hexadecimal format","value":"query hex aza"}]}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Size_must_be_greater_then_or_equals_to_3
	//Request URI: http://localhost:3000/bind/should-bind?hex=a&size=1
	//> Body: {"errors":[{"field":"size","message":"greater than or quauls to 3","value":1}]}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Size_must_be_numeric
	//Request URI: http://localhost:3000/bind/should-bind?hex=a&size=a
	//> Body: {"err":"strconv.ParseUint: parsing \"a\": invalid syntax"}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Size_must_be_greater_then_or_equals_to_3#01
	//Request URI: http://localhost:3000/bind/should-bind2?size=3
	//> Body: {"errors":[{"field":"hex","message":"required hex","value":""},{"field":"size","message":"greater than or quauls to 3","value":"3"}]}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Size_must_be_greater_then_or_equals_to_3#02
	//Request URI: http://localhost:3000/bind/should-bind2?hex=a&size=1
	//> Body: {"errors":[{"field":"size","message":"greater than or quauls to 3","value":"1"}]}
	//----------------------------------------------------------------------------------
	//=== RUN   TestShouldBind/[Query]_Size_must_be_numeric#01
	//Request URI: http://localhost:3000/bind/should-bind2?hex=a&size=a
	//> Body: {"errors":[{"field":"size","message":"size must be numeric","value":"a"}]}
	//----------------------------------------------------------------------------------
}

func putIfNotEmpty(m map[string]string, key, val string) {
	if val != "" {
		m[key] = val
	}
}

// https://github.com/gin-gonic/gin/issues/2334
// curl -XGET 'http://localhost:3500/bind-params?p1=a'
func TestQueryString(t *testing.T) {
	r := gin.Default()

	type QueryParams struct {
		P1 string `form:"p1" binding:"required,numeric"`
	}

	r.GET("/bind-params", func(c *gin.Context) {
		// check query string
		var query QueryParams
		if err := c.ShouldBindQuery(&query); err != nil {
			msg := ""
			switch err.(type) {
			case validator.ValidationErrors:
				fmt.Println("Error type :validator.ValidationErrors:", err.Error())
				e := err.(validator.ValidationErrors)[0]

				field, _ := reflect.TypeOf(&query).Elem().FieldByName(e.Field())
				fieldName, _ := field.Tag.Lookup("form")
				fmt.Println("fieldName ::", fieldName)

				//Key: Tag , Value: numeric
				//Key: ActualTag , Value: numeric
				//Key: Field , Value: P1
				//Key: Param , Value:
				//Key: Namespace , Value: QueryParams.P1
				//Key: StructField , Value: P1
				//Key: StructNamespace , Value: QueryParams.P1
				//msg: Key:P1, Value: a
				var keyValues []string
				keyValues = append(keyValues, "Tag", e.Tag())
				keyValues = append(keyValues, "ActualTag", e.ActualTag())
				keyValues = append(keyValues, "Field", e.Field())
				keyValues = append(keyValues, "Param", e.Param())
				keyValues = append(keyValues, "Namespace", e.Namespace())
				keyValues = append(keyValues, "StructField", e.StructField())
				keyValues = append(keyValues, "StructNamespace", e.StructNamespace())
				keyValues = append(keyValues, "Kind().String()", e.Kind().String())

				for i := 0; i < len(keyValues); i += 2 {
					fmt.Println("Key:", keyValues[i], ", Value:", keyValues[i+1])
				}

				msg = fmt.Sprintf("Key:%s, Value: %v", e.Field(), e.Value())
			case validator.FieldError:
				fmt.Println("Error type :validator.FieldError")
				e := err.(validator.FieldError)
				msg = fmt.Sprintf("Key:%s, Value: %v", e.Param(), e.Value())
			default:
				fmt.Println("bind query err:", err)
				msg = "cannot handle error"
			}

			fmt.Println("msg:", msg)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": msg,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"p1": query.P1,
		})
	})

	r.Run(":3500")
}
