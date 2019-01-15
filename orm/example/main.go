package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bentranter/terrible/orm"
)

// An Article is the data model representing an article.
type Article struct {
	ID        int64
	Title     string
	Text      string
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func main() {
	article := Article{}

	db := orm.Connect()

	for i := 0; i < 10; i++ {
		if err := db.Find(&article, 2); err != nil {
			log.Printf("%v", err)
		}
	}

	newArticle := Article{
		Title: "Hey!!",
		Text:  "A test.",
	}
	if err := db.Save(&newArticle); err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("%+v", newArticle)
}
