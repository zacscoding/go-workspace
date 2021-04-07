package ginexample

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func Test1(t *testing.T) {
	require := assert.New(t)
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator(zipkin.BaggagePrefix(""))
	tracer, closer := jaeger.NewTracer("sample",
		jaeger.NewConstSampler(false),
		jaeger.NewNullReporter(),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	s := NewServer(3000)
	go func() {
		if err := s.ServeHTTP(); err != nil {
			log.Fatal(err)
		}
	}()
	code, body, err := doRequest(3000, true)
	require.NoError(err)
	fmt.Println("Code: ", code, ", Body:", string(body))
}

func doRequest(port int, includeHeader bool) (int, []byte, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/root", port), nil)
	if includeHeader {
		req.Header.Set("X-Request-Id", uuid.New().String())
		req.Header.Set("X-B3-Spanid", "7cf2b60cc11b4db3")
		req.Header.Set("X-B3-Traceid", "520074fd2207a4e3")
		req.Header.Set("X-B3-Parentspanid", "520074fd2207a4e3")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}