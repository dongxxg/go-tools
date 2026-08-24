// 使用gorilla的WebSocket包，这里只是简单封装，复杂场景可以直接用原生的
// 参考：https://github.com/gorilla/websocket/blob/master/examples
package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

type WSServer struct {
	conn *websocket.Conn
}

// NewWSServer 传入read回调函数创建WebSocket server。
// write在业务代码需要的时候调用
func NewWSServer(w http.ResponseWriter, r *http.Request) (*WSServer, error) {
	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return nil, err
	}

	return &WSServer{
		conn: c,
	}, nil
}

// ReadMesssage 读取websocket消息
func (f *WSServer) ReadMesssage() (int, []byte, error) {
	if f.conn == nil {
		return -1, nil, fmt.Errorf("websocket server connection not initialized")
	}

	return f.conn.ReadMessage()
}

// WriteMessage 写入websocket消息
func (f *WSServer) WriteMessage(messageType int, data []byte) error {
	if f.conn == nil {
		return fmt.Errorf("websocket server connection not initialized")
	}

	f.conn.WriteMessage(messageType, data)
	return nil
}

// Close
func (f *WSServer) Close() {
	f.conn.Close()
	f.conn = nil
}
