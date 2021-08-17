package httpexpectexample

import (
	"github.com/gavv/httpexpect/v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBasic(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()

	e := httpexpect.New(t, srv.URL)

	res := e.GET("/hello").Expect()
	res.Status(http.StatusOK)
	res.JSON().Array().Empty()

	res = e.GET("/object").Expect()
	res.Status(http.StatusOK)
	obj := res.JSON().Object()
	obj.Value("key1").String().Equal("value1")
	obj.Value("key2").Array().Equal([]interface{}{"value21", "value22"})
}

func newHandler() http.Handler {
	h := &handler{}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", h.handleHello)
	mux.HandleFunc("/object", h.handleObject)

	return mux
}

type handler struct {
}

func (h *handler) handleHello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[]`))
}

func (h *handler) handleObject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`
	{
		"key1": "value1",
		"key2": [
			"value21", "value22"
		],
		"key3": {
			"key33": "value331"
		}
	}`))
}
