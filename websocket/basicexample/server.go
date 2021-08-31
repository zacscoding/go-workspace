package main

import (
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var pongWait = time.Second * 5

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func echo(w http.ResponseWriter, r *http.Request) {
	requestLog(r)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		serverLogf("failed to update. err: %v", err)
		return
	}
	defer c.Close()

	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPingHandler(func(appData string) error {
		serverLogf("Handle ping.. : %s", appData)
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		mtype, msg, err := c.ReadMessage()
		if err != nil {
			serverLogf("failed to read message. err: %v", err)
			break
		}

		serverLogf("recv: type: %d, msg: %s", mtype, msg)
		err = c.WriteMessage(mtype, msg)
		if err != nil {
			serverLogf("failed to write message. err: %v", err)
		}
	}
}

func serverLogf(format string, v ...interface{}) {
	log.Printf("[Server]"+format, v...)
}

func requestLog(r *http.Request) {
	serverLogf("# Request")
	serverLogf("URL: %s, Method: %s", r.RequestURI, r.Method)
	for k, v := range r.Header {
		serverLogf("Header key: %s, values: %s", k, strings.Join(v, ","))
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
