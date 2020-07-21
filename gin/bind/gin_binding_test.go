package bind

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"testing"
)

// WTD
// (1) query string name / uri path name
// (2) numeric binding

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
