package client

import (
	"Websocket/dao"
	"log"
	"sync"
)

type Room struct {
	Name           string
	Hub            *Hub
	Clients        map[*Client]bool
	HistoryMessage [][]byte
	Mutex          sync.Mutex
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func NewRoom(name string) *Room {
	r := &Room{
		Name:    name,
		Hub:     NewHub(),
		Clients: make(map[*Client]bool),
	}
	r.LoadHistory() // 加载历史记录
	return r
}

func (r *Room) Join(cli *Client) {
	r.Clients[cli] = true
	cli.Room = r
}

func (r *Room) Leave(cli *Client) {
	// 移除客户端
	delete(r.Clients, cli)
	if len(r.Clients) == 0 {
		Rooms.Delete(r.Name)
	}
}

func (r *Room) Broadcast(msg []byte) {
	// 添加消息到历史记录
	r.Mutex.Lock()
	r.HistoryMessage = append(r.HistoryMessage, msg)
	dao.SaveMessage(r.Name, string(msg))

	r.Mutex.Unlock()

	// 广播消息给所有客户端
	for client := range r.Clients {
		select {
		case client.Send <- msg:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

func (r *Room) Run() {
	r.Hub.Run()
}

func (r *Room) LoadHistory() { // 加载历史记录
	messages, err := dao.LoadHistory(r.Name)
	if err != nil {
		log.Println("failed to load history message:", err)
		return
	}

	for _, msg := range messages {
		r.HistoryMessage = append(r.HistoryMessage, []byte(msg))
	}
}

func NewHub() *Hub { // 创建一个新的Hub
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() { // 运行Hub
	for {
		select {
		case cli := <-h.Register:
			h.Clients[cli] = true
		case cli := <-h.Unregister:
			if _, ok := h.Clients[cli]; ok {
				delete(h.Clients, cli)
				close(cli.Send)
			}
		case msg := <-h.Broadcast:
			for cli := range h.Clients {
				select {
				case cli.Send <- msg:
				default:
					close(cli.Send)
					delete(h.Clients, cli)
				}
			}
		}
	}
}
