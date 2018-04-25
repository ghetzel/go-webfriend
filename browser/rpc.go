package browser

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
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
	sendlock            sync.Mutex
	closing             bool
}

type RpcError struct {
	Code    int
	Message string
}

func (self *RpcError) Error() string {
	return fmt.Sprintf("code %d: %v", self.Code, self.Message)
}

type RpcMessage struct {
	ID     int64                  `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
	Error  map[string]interface{} `json:"error,omitempty"`
}

func (self *RpcMessage) P() *maputil.Map {
	return maputil.M(self.Params)
}

func (self *RpcMessage) R() *maputil.Map {
	return maputil.M(self.Result)
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

func (self *RPC) SynthesizeEvent(message RpcMessage) {
	self.recv <- &message
}

func (self *RPC) startReading() {
	for {
		if self.closing {
			return
		}

		message := &RpcMessage{}

		if _, data, err := self.conn.ReadMessage(); err == nil {
			if err := json.Unmarshal(data, message); err == nil {
				waitingForId := atomic.LoadInt64(&self.waitingForMessageId)

				// if we just read a message another
				if waitingForId > 0 && int64(message.ID) == waitingForId {
					log.Debugf("[rpc] REPLY %d", message.ID)
					self.reply <- message
				} else {
					self.recv <- message
				}
			} else {
				log.Errorf("Failed to decode RPC message: %v", err)
				return
			}
		} else if self.closing {
			return
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
	if self.closing {
		return nil, fmt.Errorf("Cannot send, connection is closing...")
	}

	self.sendlock.Lock()
	defer self.sendlock.Unlock()

	mid := atomic.AddInt64(&self.messageId, 1)
	message.ID = mid
	waitForReply := (timeout > 0)

	if waitForReply {
		atomic.StoreInt64(&self.waitingForMessageId, mid)
	}

	if err := self.conn.WriteJSON(message); err == nil {
		log.Debugf("[rpc] WROTE: %v", message)

		if waitForReply {
			select {
			case reply := <-self.reply:
				atomic.StoreInt64(&self.waitingForMessageId, 0)
				return reply, nil

			case <-time.After(timeout):
				return nil, fmt.Errorf("Timed out waiting for reply to message %d", mid)
			}
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func (self *RPC) Close() error {
	if !self.closing {
		log.Debug("[rpc] Closing RPC connection")
		self.closing = true
		close(self.recv)
	}

	return self.conn.Close()
}
