package websocket

import (
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn *websocket.Conn
}

// NewClient 新建websocket客户端
func NewClient(serverAddr, path string) (*WSClient, error) {
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: path}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &WSClient{
		conn: c,
	}, nil
}

// ReadMesssage 读取websocket消息
func (f *WSClient) ReadMesssage() (int, []byte, error) {
	if f.conn == nil {
		return -1, nil, fmt.Errorf("websocket connection not initialized")
	}

	return f.conn.ReadMessage()
}

// 写入websocket
func (f *WSClient) WriteMessage(messageType int, data []byte) error {
	if f.conn == nil {
		return fmt.Errorf("websocket server not initialized")
	}

	f.conn.WriteMessage(messageType, data)
	return nil
}

// Close
func (f *WSClient) Close() {
	f.conn.Close()
	f.conn = nil
}
