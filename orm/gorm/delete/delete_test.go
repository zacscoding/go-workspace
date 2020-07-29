package delete

import (
	"fmt"
	"testing"
)

func TestDeleteAccount(t *testing.T) {
	article1 := Article{
		Title: "title1",
	}
	article2 := Article{
		Title: "title1",
	}
	acc1 := &Account{
		Name:     "acc1",
		Articles: []Article{article1, article2},
	}

	err := db.Save(acc1).Error
	if err != nil {
		panic(err)
	}

	// account's delete is rollback if provide true flag
	err = DeleteAccount(acc1.ID, true)
	if err == nil {
		panic("exception err")
	}
	fmt.Println("Error >", err.Error())
	fmt.Println("## After deleted with fail")
	displayAccounts()
	displayArticles()

	// success to delete account, articles
	err = DeleteAccount(acc1.ID, false)
	if err != nil {
		panic(err)
	}
	fmt.Println("## After deleted")
	displayAccounts()
	displayArticles()
}

func displayAccounts() {
	var accounts []Account
	err := db.Preload("Articles").Find(&accounts).Error
	if err != nil {
		panic(err)
	}
	fmt.Println("Total accounts :", len(accounts))
	for _, account := range accounts {
		fmt.Println("Acc >> id:", account.ID, ", name:", account.Name)
		for _, article := range account.Articles {
			fmt.Println(" Article >> id:", article.ID, ", title:", article.Title)
		}
		fmt.Println("---------------------------------------")
	}
}

func displayArticles() {
	var articles []Article
	err := db.Find(&articles).Error
	if err != nil {
		panic(err)
	}
	fmt.Println("Total article :", len(articles))
	for _, article := range articles {
		fmt.Println("Article >> id:", article.ID, ", title:", article.Title)
	}
}
