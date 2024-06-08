package resp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Stats   int
	Message string
}

var (
	InternetServerError = Response{
		Stats:   500,
		Message: "Internet Server Error",
	}

	RoomNotFound = Response{
		Stats:   40000,
		Message: "Room not found",
	}

	Timeout = Response{
		Stats:   40001,
		Message: "Chat Time has not arrive",
	}
)

func InternetError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, InternetServerError)
}

func RoomExist(c *gin.Context) {
	c.JSON(http.StatusBadRequest, RoomNotFound)
}

func TimeNotArrive(c *gin.Context) {
	c.JSON(http.StatusBadRequest, Timeout)
}
