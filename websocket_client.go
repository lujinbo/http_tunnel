package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// 代理服务器地址
	proxyURL, _ := url.Parse("http://192.168.3.221:8780")
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Proxy: http.ProxyURL(proxyURL), //挂代理
	}

	//自己搭建一个websocket服务，监听127.0.0.1:7777，且与代理服务在一台机器上
	wsURL := "ws://127.0.0.1:7777/echo"
	wsConn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}
	defer wsConn.Close()

	// 发送WebSocket消息
	if err := wsConn.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!")); err != nil {
		fmt.Println("Error writing to WebSocket:", err)
		return
	}

	// 接收WebSocket消息
	ticker := time.NewTicker(time.Second * 3)
	go func() {
		for {
			select {
			case <-ticker.C:
				wsConn.WriteMessage(1, []byte("北京时间："+time.Now().Format("2006-01-02 15:04:05")))
			}
		}
	}()

	for {
		_, mdata, err := wsConn.ReadMessage()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from WebSocket:", err)
				return
			}
		}
		fmt.Printf("Received message from WebSocket server: %s\n", string(mdata))
	}
}
