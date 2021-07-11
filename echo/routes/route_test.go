package routes

import (
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"testing"
	"time"
)

type ListOpt struct {
	Size   uint   `query:"size"`
	Cursor string `query:"cursor"`
}

type User struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var (
	users = []*User{
		{
			Name:      "user1",
			Email:     "user1@gmail.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:      "user1",
			Email:     "user1@gmail.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
)

func TestEchoRoutes(t *testing.T) {
	jaegertracing.New()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	userGroup := e.Group("/api/user")
	userGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			log.Println("UserGroup middleware is called")
			if err := next(ctx); err != nil {
				ctx.Error(err)
			}
			return nil
		}
	})
	userGroup.GET("", func(ctx echo.Context) error {
		//size := ctx.QueryParam("size")
		//cursor := ctx.QueryParam("cursor")
		opt := new(ListOpt)
		if err := ctx.Bind(opt); err != nil {
			return err
		}
		log.Printf("%s is called > size: %d, cursor:%s", ctx.Path(), opt.Size, opt.Cursor)
		ctx.JSON(http.StatusOK, echo.Map{
			"users": users,
		})
		return nil
	})
	userGroup.GET("/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")
		log.Printf("%s is called > id: %s", ctx.Path(), id)

		if id == "1" {
			ctx.JSON(http.StatusNotFound, echo.Map{
				"message": "user not found",
			})
			return nil
		}
		if id == "2" {
			return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
				"message": "invalid user id",
			})
		}
		ctx.JSON(http.StatusOK, users[0])
		return nil
	})
	userGroup.GET("/search", func(ctx echo.Context) error {
		log.Printf("%s is called", ctx.Path())
		ctx.JSON(http.StatusOK, users)
		return nil
	})

	articleGroup := e.Group("/api/article")
	articleGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			log.Println("ArticleGroup middleware is called")
			next(ctx)
			return nil
		}
	})
	articleGroup.GET("", func(ctx echo.Context) error {
		ctx.JSON(http.StatusOK, echo.Map{})
		return nil
	})

	if err := e.Start(":8800"); err != nil {
		log.Fatal(err)
	}
}
