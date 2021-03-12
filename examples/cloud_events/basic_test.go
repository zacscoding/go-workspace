package cloud_events

import (
	"context"
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpServerAndClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Println("------------------------------")
		log.Println("---- Headers")
		for k, values := range req.Header {
			log.Printf("Key:%s, Values:%s", k, strings.Join(values, ","))
		}
		log.Println("------------------------------")
		log.Println("---- Request Body")
		body, _ := ioutil.ReadAll(req.Body)
		log.Println("Body:", string(body))
	}))
	defer server.Close()

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// events
	e := cloudevents.NewEvent()
	e.SetSource("example/uri")
	e.SetType("example.type")
	e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"message": "Hello CloudEvents :)",
	})

	ctx := cloudevents.ContextWithTarget(context.Background(), server.URL)
	if result := c.Send(ctx, e); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send. err:%v", result)
	}
	// Output
	//2021/03/13 01:48:00 ------------------------------
	//2021/03/13 01:48:00 ---- Headers
	//2021/03/13 01:48:00 Key:Content-Length, Values:34
	//2021/03/13 01:48:00 Key:Ce-Id, Values:44cc66f4-979f-478d-88ab-7ea30552a642
	//2021/03/13 01:48:00 Key:Ce-Time, Values:2021-03-12T16:48:00.768636Z
	//2021/03/13 01:48:00 Key:Content-Type, Values:application/json
	//2021/03/13 01:48:00 Key:Accept-Encoding, Values:gzip
	//2021/03/13 01:48:00 Key:User-Agent, Values:Go-http-client/1.1
	//2021/03/13 01:48:00 Key:Ce-Source, Values:example/uri
	//2021/03/13 01:48:00 Key:Ce-Specversion, Values:1.0
	//2021/03/13 01:48:00 Key:Ce-Type, Values:example.type
	//2021/03/13 01:48:00 ------------------------------
	//2021/03/13 01:48:00 ---- Request Body
	//2021/03/13 01:48:00 Body: {"message":"Hello CloudEvents :)"}
}

func TestSerializeAndDeserialize(t *testing.T) {
	e := cloudevents.NewEvent()
	e.SetID("example-id")
	e.SetSource("example/uri")
	e.SetType("example.type")
	e.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"message": "Hello CloudEvents :)",
	})
	bytes, _ := json.Marshal(e)
	log.Println("------------------------------")
	log.Println("Marshal: ", string(bytes))
	log.Println("------------------------------")
	log.Println("Unmarshal")
	read := cloudevents.NewEvent()
	json.Unmarshal(bytes, &read)
	log.Println("ID:", read.ID())
	log.Println("Source:", read.Source())
	log.Println("Type:", read.Type())
	log.Println("Data:", string(read.Data()))
	log.Println("DataContentType:", read.DataContentType())
	//2021/03/13 01:54:37 ------------------------------
	//2021/03/13 01:54:37 Marshal:  {"data":{"message":"Hello CloudEvents :)"},"datacontenttype":"application/json","id":"example-id","source":"example/uri","specversion":"1.0","type":"example.type"}
	//2021/03/13 01:54:37 ------------------------------
	//2021/03/13 01:54:37 Unmarshal
	//2021/03/13 01:54:37 ID: example-id
	//2021/03/13 01:54:37 Source: example/uri
	//2021/03/13 01:54:37 Type: example.type
	//2021/03/13 01:54:37 Data: {"message":"Hello CloudEvents :)"}
	//2021/03/13 01:54:37 DataContentType: application/json
}
