package bind

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

var cgteValidator validator.Func = func(fl validator.FieldLevel) bool {
	v := fl.Field().Interface()
	switch v.(type) {
	case Param:
		p := v.(Param)
		intVal, err := p.Int()
		if err != nil {
			panic(fmt.Sprintf("Param must be use with numeric. err:%v", err))
		}
		param, err := strconv.Atoi(fl.Param())
		if err != nil {
			panic(fmt.Sprintf("Param must be numeric. but:%s. err:%v", fl.Param(), err))
		}
		return intVal >= param
	default:
		panic("cgte validator must use with string type")
	}
}

type Param string

func (p Param) Int() (int, error) {
	return strconv.Atoi(string(p))
}

func (p Param) MustInt() int {
	ret, err := p.Int()
	if err == nil {
		return ret
	}
	panic("Err:" + err.Error())
}

func (p Param) Uint(base, bitSize int) (uint, error) {
	ret, err := strconv.ParseUint(string(p), base, bitSize)
	if err != nil {
		return 0, err
	}
	return uint(ret), nil
}

func (p Param) MustUint(base, bitSize int) (uint, error) {
	ret, err := p.Uint(base, bitSize)
	if err == nil {
		return ret, nil
	}
	panic("Err:" + err.Error())
}

func StartGinServer() {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("cgte", cgteValidator)
	}

	handleErr := func(c *gin.Context, obj interface{}, tag string, err error) {
		switch err.(type) {
		case validator.ValidationErrors:
			var errs []gin.H
			vErrs := err.(validator.ValidationErrors)
			e := reflect.TypeOf(obj).Elem()
			for _, vErr := range vErrs {
				field, _ := e.FieldByName(vErr.Field())

				tagName, _ := field.Tag.Lookup(tag)
				value := vErr.Value()
				var message string
				switch vErr.ActualTag() {
				case "required":
					message = fmt.Sprintf("required %s", tagName)
				case "hexadecimal":
					message = fmt.Sprintf("required hexadecimal format")
				case "gte":
					message = fmt.Sprintf("greater than or quauls to %s", vErr.Param())
				case "cgte":
					message = fmt.Sprintf("greater than or quauls to %s", vErr.Param())
				case "numeric":
					message = fmt.Sprintf("%s must be numeric", tagName)
				default:
					message = err.Error()
				}

				errs = append(errs, gin.H{
					"field":   tagName,
					"value":   value,
					"message": message,
				})
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errors": errs,
			})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"err": err.Error(),
			})
		}
	}

	e.GET("bind/should-bind", func(c *gin.Context) {
		type QueryParameter struct {
			Hex  string `form:"hex" binding:"required,hexadecimal"`
			Size uint   `form:"size" binding:"omitempty,gte=3"`
		}
		var query QueryParameter

		if err := c.ShouldBind(&query); err != nil {
			handleErr(c, &query, "form", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query": gin.H{
				"hex":  query.Hex,
				"size": query.Size,
			},
		})
	})

	e.GET("bind/should-bind2", func(c *gin.Context) {
		type QueryParameter struct {
			Hex  string `form:"hex" binding:"required,hexadecimal"`
			Size Param  `form:"size" binding:"omitempty,numeric,cgte=3"`
		}
		var query QueryParameter
		if err := c.ShouldBind(&query); err != nil {
			handleErr(c, &query, "form", err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"query": gin.H{
				"hex":  query.Hex,
				"size": query.Size.MustInt(),
			},
		})
	})

	if err := e.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}

// References
// https://github.com/gin-gonic/gin/issues/430
//
