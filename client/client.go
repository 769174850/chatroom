package client

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var Upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	Rooms = sync.Map{}
	Lock  = sync.Mutex{}
)

type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	Name    string
	Send    chan []byte
	Room    *Room
	closeMu sync.Mutex
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Printf("user %s exit the room %s\n", c.Name, c.Room.Name)
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("write msg failed, err:", err)
				return
			}
			w.Write(msg)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				log.Println("write msg close failed, err:", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("ticker write msg failed, err:", err)
				return
			}
		case <-Interrupt:
			c.Conn.WriteMessage(websocket.CloseMessage, []byte("client closed"))
			return
		}
	}
}

func (c *Client) Read() {
	defer func() {
		log.Printf("user %s exit the room %s\n", c.Name, c.Room.Name)
		c.Hub.Unregister <- c
		Rooms.Delete(c.Name)
		c.Room.Leave(c)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		msgType, msgByte, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read msg failed, err:", err)
			}
			// Check if the connection is already closed
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				log.Println("connection already closed")
				return
			}
			c.Reconnect(5, 5*time.Second, 30*time.Second) // 尝试重连，最大重试次数为 5 次
			break
		}
		switch msgType {
		case websocket.TextMessage:
			msg := []byte(fmt.Sprintf("%s %s说:%s",
				time.Now().Format("01/02 03:04"), c.Name, string(msgByte)))
			Rooms.Range(func(key, value any) bool {
				value.(chan []byte) <- msg
				return true
			})
		default:
			log.Println("receive don't know msg type is ", msgType)
			continue
		}
		msgByte = bytes.TrimSpace(bytes.Replace(msgByte, newline, space, -1))
		c.Hub.Broadcast <- msgByte
	}
}

func (c *Client) Reconnect(maxRetries int, initialDelay time.Duration, maxDelay time.Duration) {
	retries := 0
	delay := initialDelay
	for retries < maxRetries {
		log.Println("Attempting to reconnect...")
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/chatroom/"+c.Room.Name, nil)
		if err != nil {
			log.Printf("Reconnect failed, retrying in %s... (%d/%d)", delay, retries+1, maxRetries)
			time.Sleep(delay)
			retries++
			// 计算下一次重连的延迟时间，加入随机因子
			randomFactor := time.Duration(rand.Int63n(int64(delay / 2)))
			delay = delay*2 + randomFactor
			if delay > maxDelay {
				delay = maxDelay
			}
			continue
		}

		c.Conn = conn
		go c.Write()
		go c.Read()
		log.Println("Reconnected successfully")
		return
	}

	log.Println("Max retries reached, could not reconnect")
}

func ServeWs(room *Room, hub *Hub, w http.ResponseWriter, r *http.Request) (err error) {
	conn, err := Upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade to websocket failed, err:", err)
		return
	}

	rand.Seed(time.Now().UnixMicro())

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		Name: "user" + strconv.Itoa(rand.Intn(100000)+10000),
	}

	Lock.Lock()
	Rooms.Store(client.Name, client.Send)
	Lock.Unlock()

	room.Join(client)

	go room.Run()

	go client.Write()
	go client.Read()

	return
}
