package middlewares

import "testing"

func TestServer(t *testing.T) {
	if err := StartServer(":8900"); err != nil {
		t.Log(err)
	}
}
