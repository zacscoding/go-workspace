package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"time"
)

// jwtCustomClaims are custom claims extending default ones.
type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

func StartServer(addr string) error {
	e := echo.New()
	e.Use(middleware.Recover())

	e.POST("/signin", handleSignIn)
	e.GET("/public", handlePublic)
	r := e.Group("/auth")
	// Configure middleware with the custom claims type
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
		BeforeFunc: func(ctx echo.Context) {
			log.Println("[Echo] BeforeFunc is called")
		},
		SuccessHandler: func(ctx echo.Context) {
			user := ctx.Get("user").(*jwt.Token)
			claims := user.Claims.(*jwtCustomClaims)
			log.Printf("[Echo] SuccessHandler is called. name: %s", claims.Name)
		},
		ErrorHandler: func(err error) error {
			log.Printf("[Echo] ErrorHandler is called. err: %v", err)
			return echo.ErrUnauthorized
		},
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("", handleAuth)
	return e.Start(addr)
}

func handleSignIn(ctx echo.Context) error {
	username := ctx.FormValue("username")
	password := ctx.FormValue("password")

	if username != "zac" || password != "coding" {
		return echo.ErrUnauthorized
	}

	claims := &jwtCustomClaims{
		Name:  username,
		Admin: true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func handlePublic(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, echo.Map{
		"message": "Hello :)",
	})
}

func handleAuth(ctx echo.Context) error {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	return ctx.JSON(http.StatusOK, echo.Map{
		"name":    claims.Name,
		"isAdmin": claims.Admin,
	})
}
