package httpexpectexample

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

type GinHandler struct{}

func TestGin(t *testing.T) {
	h := &GinHandler{}
	gin.SetMode(gin.TestMode)
	e := gin.Default()
	e.GET("/object", h.handleObject)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)

	e.ServeHTTP(rec, req)

	expect := httpexpect.NewResponse(t, rec.Result())
	expect.Status(http.StatusOK)
	expect.JSON().Path("$.message").String().ContainsFold("hello")
}

func TestGinObject(t *testing.T) {
	h := &GinHandler{}
	gin.SetMode(gin.TestMode)
	e := gin.Default()
	e.GET("/object", h.handleObject)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/object", nil)

	e.ServeHTTP(rec, req)

	expect := httpexpect.NewResponse(t, rec.Result())
	expect.Status(http.StatusOK)
	obj := expect.JSON().Object()
	obj.Value("key1").String().Equal("value1")
	obj.Value("key2").Array().Equal([]interface{}{"value21", "value22"})
}

func (h *GinHandler) handleHello(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Hello :)",
	})
}

func (h *GinHandler) handleObject(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"key1": "value1",
		"key2": []string{"value21", "value22"},
		"key3": gin.H{
			"key33": "value331",
		},
	})
}
