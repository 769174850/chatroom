package operate

import (
	"Websocket/client"
	"Websocket/resp"
	"github.com/gin-gonic/gin"
	"time"
)

func Chat(c *gin.Context) {
	// 判断是否在规定时间内
	NowTime := time.Now()
	startTime := time.Date(NowTime.Year(), NowTime.Month(), NowTime.Day(), 23, 0, 0, 0, NowTime.Location())
	endTime := time.Date(NowTime.Year(), NowTime.Month(), NowTime.Day(), 6, 0, 0, 0, NowTime.Location())

	if NowTime.After(startTime) || NowTime.Before(endTime) {
		resp.TimeNotArrive(c)
		return
	}

	// 获取URL参数
	id := c.Param("id")
	if client.RoomExist(id) == true {
		resp.RoomExist(c)
		return
	}

	// 解析命令行参数
	hub := client.NewHub()
	go hub.Run()

	// 创建房间
	room := client.NewRoom(id)

	//设置客户端，在连接失败时尝试重新连接
	err := client.ServeWs(room, hub, c.Writer, c.Request)
	if err != nil {
		resp.InternetError(c)
		return
	}
}
