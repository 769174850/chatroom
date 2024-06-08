package dao

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

type Message struct {
	ID        uint `gorm:"primaryKey"`
	RoomName  string
	Message   string
	CreatedAt time.Time
}

func init() {
	_, err := InitDB()
	if err != nil {
		log.Println("failed to init sql:", err)
		return
	}
}

func InitDB() (db *gorm.DB, err error) {
	var dns = "root:123456@tcp(127.0.0.1:3306)/chatroom?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		log.Println("failed to connect sql:", err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Println("failed to get sqlDB:", err)
	}
	sqlDB.SetMaxOpenConns(10) //设置最大的连接数
	sqlDB.SetMaxIdleConns(5)  //设置最大空闲数

	err = sqlDB.Ping() //检查数据库链接问题
	if err != nil {
		log.Println(err)
	}

	DB = db

	err = db.AutoMigrate(&Message{})
	if err != nil {
		log.Println("failed to create table :", err)
	}

	return
}

func SaveMessage(roomName, message string) {
	msg := Message{RoomName: roomName, Message: message}
	if err := DB.Create(&msg).Error; err != nil {
		log.Println("failed to save message:", err)
	}
}

func LoadHistory(roomName string) ([]string, error) {
	var messages []Message
	if err := DB.Where("room_name = ?", roomName).Find(&messages).Error; err != nil {
		return nil, err
	}

	var history []string
	for _, msg := range messages {
		history = append(history, msg.Message)
	}
	return history, nil
}
