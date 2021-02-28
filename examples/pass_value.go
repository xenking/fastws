// +build ignore
package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/xenking/fastws"
)

var userCount = int32(0)

func main() {
	router := fasthttprouter.New()
	router.GET("/", rootHandler)
	router.GET("/ws", func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue("user", atomic.AddInt32(&userCount, 1))
		fastws.Upgrade(wsHandler)(ctx)
	})

	server := fasthttp.Server{
		Handler: router.Handler,
	}
	go server.ListenAndServe(":8080")

	fmt.Println("Visit http://localhost:8080")

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	signal.Stop(sigCh)
	signal.Reset(os.Interrupt)
	server.Shutdown()
}

func wsHandler(conn *fastws.Conn) {
	fmt.Printf("Opened connection\n")

	user := conn.UserValue("user").(int32)

	fmt.Fprintf(conn, "Hello user %d!", user)

	var msg []byte
	var err error
	for {
		_, msg, err = conn.ReadMessage(msg[:0])
		if err != nil {
			if err != fastws.EOF {
				fmt.Fprintf(os.Stderr, "error reading message: %s\n", err)
			}
			break
		}
		time.Sleep(time.Second)

		_, err = conn.Write(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing message: %s\n", err)
			break
		}
	}

	fmt.Printf("Closed connection\n")
}

func rootHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	fmt.Fprintln(ctx, `<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8"/>
    <title>Sample of websocket with Golang</title>
  </head>
  <body>
		<div id="text"></div>
    <script>
      var ws = new WebSocket("ws://localhost:8080/ws");
      ws.onmessage = function(e) {
				var d = document.createElement("div");
        d.innerHTML = e.data;
				ws.send(e.data);
        document.getElementById("text").appendChild(d);
      }
			ws.onclose = function(e){
				var d = document.createElement("div");
				d.innerHTML = "CLOSED";
        document.getElementById("text").appendChild(d);
			}
    </script>
  </body>
</html>`)
}
