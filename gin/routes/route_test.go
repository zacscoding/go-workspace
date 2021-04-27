package routes

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"testing"
)

func Test1(t *testing.T) {
	e := gin.Default()

	e.POST("/metadata/upload", handleRequest)
	e.GET("/metadata/asset/:id", handleRequest)
	e.GET("/metadata/:id", handleRequest)

	if err := e.Run("8800"); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(ctx *gin.Context) {
	log.Printf("%s is called..", ctx.FullPath())
	ctx.JSON(http.StatusOK, gin.H{
		"message": "hello",
	})
}
