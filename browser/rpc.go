package browser

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/gorilla/websocket"
)

var MaxUnreadEvents = 1024
var DefaultReplyTimeout = 10 * time.Second

type RPC struct {
	URL                 string
	conn                *websocket.Conn
	messageId           int64
	waitingForMessageId int64
	recv                chan *RpcMessage
	reply               chan *RpcMessage
}

type RpcError struct {
	Code    int
	Message string
}

type RpcMessage struct {
	ID     int64                  `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
	Result map[string]interface{} `json:"results,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

func (self *RpcMessage) Err() *RpcError {
	if self.Error != `` {
		return &RpcError{
			Message: self.Error,
		}
	} else {
		return nil
	}
}

func (self *RpcMessage) String() string {
	if data, err := json.Marshal(self); err == nil {
		return string(data)
	} else {
		return fmt.Sprintf("ERR<%v>", err)
	}
}

func NewRPC(wsUrl string) (*RPC, error) {
	rpc := &RPC{
		URL:   wsUrl,
		recv:  make(chan *RpcMessage),
		reply: make(chan *RpcMessage),
	}

	if conn, _, err := websocket.DefaultDialer.Dial(rpc.URL, nil); err == nil {
		rpc.conn = conn
		go rpc.startReading()

		return rpc, nil
	} else {
		return nil, err
	}
}

func (self *RPC) startReading() {
	for {
		message := &RpcMessage{}

		if _, data, err := self.conn.ReadMessage(); err == nil {

			if err := json.Unmarshal(data, message); err == nil {
				if id := message.ID; id > 0 {
					log.Debugf("[rpc] READ: %v", id)
				}

				// if we just read a message another
				if id := atomic.LoadInt64(&self.waitingForMessageId); id > 0 && int64(message.ID) == id {
					self.reply <- message
				} else {
					self.recv <- message
				}
			} else {
				log.Errorf("Failed to decode RPC message: %v", err)
				return
			}
		} else {
			log.Errorf("Failed to read from RPC: %v", err)
			return
		}
	}
}

func (self *RPC) Messages() <-chan *RpcMessage {
	return self.recv
}

func (self *RPC) Call(method string, params map[string]interface{}, timeout time.Duration) (*RpcMessage, error) {
	message := &RpcMessage{
		Method: method,
		Params: params,
	}

	return self.Send(message, timeout)
}

func (self *RPC) CallAsync(method string, params map[string]interface{}) error {
	message := &RpcMessage{
		Method: method,
		Params: params,
	}

	_, err := self.Send(message, 0)
	return err
}

func (self *RPC) Send(message *RpcMessage, timeout time.Duration) (*RpcMessage, error) {
	message.ID = atomic.AddInt64(&self.messageId, 1)
	waitForReply := (timeout > 0)

	if waitForReply {
		atomic.StoreInt64(&self.waitingForMessageId, message.ID)
	}

	if err := self.conn.WriteJSON(message); err == nil {
		log.Debugf("[rpc] WROTE: %v", message)

		if waitForReply {
			select {
			case reply := <-self.reply:
				atomic.StoreInt64(&self.waitingForMessageId, 0)
				return reply, nil
			case <-time.After(timeout):
				return nil, fmt.Errorf("Timed out waiting for reply to message %d", message.ID)
			}
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func (self *RPC) Close() error {
	return self.conn.Close()
}
