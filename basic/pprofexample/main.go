package main

import (
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	e := gin.Default()
	pprof.Register(e)

	e.GET("/goroutine", func(ctx *gin.Context) {
		log.Println("# Start go routines")
		workers, err := extractIntQueryParams(ctx, "workers")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		duration := ctx.Query("duration")
		d := time.Second
		if duration != "" {
			d, err = time.ParseDuration(duration)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
		}

		for i := 0; i < workers; i++ {
			go func() {
				time.Sleep(d)
			}()
		}
		ctx.JSON(http.StatusOK, gin.H{
			"workers":  workers,
			"duration": d.String(),
		})
	})

	if err := e.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func extractIntQueryParams(ctx *gin.Context, key string) (int, error) {
	val := ctx.Query(key)
	if val == "" {
		return 0, fmt.Errorf("no query params: %s", key)
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}
