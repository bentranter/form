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
	db := orm.Connect()

	// SELECT * FROM articles
	articles := make([]*Article, 0)
	if err := db.All(&articles); err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Printf("Found %d articles\n", len(articles))

	article := Article{}
	// SELECT * FROM articles WHERE id = $1
	if err := db.Find(&article, 2); err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Printf("Found article: %+v\n", article)

	newArticle := Article{
		Title: "Saving test",
		Text:  "A test.",
	}

	// INSERT INTO articles (created_at,text,title,updated_at) VALUES ($1,$2,$3,$4) RETURNING *
	if err := db.Save(&newArticle); err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("New article: %+v\n", newArticle)

	// UPDATE articles SET text = $1, title = $2, updated_at = $3 WHERE id = $4 RETURNING *
	newArticle.Title = "Updating test"
	if err := db.Update(&newArticle); err != nil {
		log.Printf("%v", err)
	}
	fmt.Printf("Updated article: %+v\n", newArticle)

	// DELETE
	if err := db.Destroy(&newArticle); err != nil {
		log.Printf("%v", err)
	}
	fmt.Println("Deleted article")
}
