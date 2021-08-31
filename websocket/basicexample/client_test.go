package main

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/url"
	"testing"
	"time"
)

func TestWSClient(t *testing.T) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/echo"}
	t.Logf("connection to %s", u.String())

	c, res, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Logf("failed to connect. err: %v", err)
		return
	}
	defer c.Close()
	defer res.Body.Close()
	respBody, _ := ioutil.ReadAll(res.Body)
	t.Logf("resp code: %d, body: %s", res.StatusCode, string(respBody))
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				t.Log("try to send a ping message")
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					t.Logf("failed to write ping message: %v", err)
					return
				}
			}
		}
	}()
	go func() {
		for {
			mtype, payload, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					t.Logf("failed to read message: %v", err)
				}
				t.Logf("closed connection")
				return
			}
			t.Logf("recv > type: %d, message: %s", mtype, string(payload))
		}
	}()
	if err := c.WriteMessage(websocket.TextMessage, []byte("Hello")); err != nil {
		t.Logf("failed to write a message: %v", err)
		return
	}
	time.Sleep(pongWait + time.Second)
	t.Logf("Terminate client..")
}
