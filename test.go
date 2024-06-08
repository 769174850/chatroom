package main

import (
	"Websocket/dao"
	"log"
)

func main() {
	// 初始化数据库连接
	_, err := dao.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 保存聊天记录
	saveMessageTest()

	// 加载聊天记录
	loadHistoryTest()
}

func saveMessageTest() {
	roomName := "testRoom"
	message := "Hello, World!"

	// 保存消息到数据库
	dao.SaveMessage(roomName, message)
	log.Println("Message saved successfully.")
}

func loadHistoryTest() {
	roomName := "testRoom"

	// 加载房间的聊天记录
	history, err := dao.LoadHistory(roomName)
	if err != nil {
		log.Fatalf("Failed to load chat history: %v", err)
	}

	log.Println("Chat history for room", roomName, ":")
	for _, msg := range history {
		log.Println(msg)
	}
}
