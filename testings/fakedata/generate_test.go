package fakedata

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"log"
	"testing"
	"time"
)

type User struct {
	Fistname string `json:"firstName"`
	LastName string `json:"lastName"`
	Email    string `json:"email"`
	Company  string `json:"company"`
}

func TestGenerateUserFake(t *testing.T) {
	var (
		count = 5
		users []User
	)

	gofakeit.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		u := User{
			Fistname: gofakeit.FirstName(),
			LastName: gofakeit.LastName(),
			Email:    gofakeit.Email(),
			Company:  gofakeit.Company(),
		}
		users = append(users, u)
	}

	b, _ := json.MarshalIndent(&users, "", "    ")
	log.Println("## Users >\n", string(b))
}

type Article struct {
	Title       string `fake:"{phrase}"`
	Description string
	Body        string `fake:"skip"`
	Author      string `fake:"{name}"`
}

func TestGenerateArticleFake(t *testing.T) {
	var (
		count    = 5
		articles []Article
	)

	gofakeit.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		var a Article
		gofakeit.Struct(&a)
		articles = append(articles, a)
	}

	b, _ := json.MarshalIndent(&articles, "", "    ")
	log.Println("## Articles >\n", string(b))
}
