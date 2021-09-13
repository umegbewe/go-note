package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3" // go-sqlite3 library
	pusher "github.com/pusher/pusher-http-go"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db := initialiseDatabase("./db/posts.db")
	migrateDatabase(db)

	e.File("/", "frontend/index.html")
	e.GET("/posts", getPosts(db))
	e.POST("/posts", savePost(db))
	e.Logger.Fatal(e.Start(":8081"))
}

var client = pusher.Client{
	AppID:   "PUSHER_APP_ID",
	Key:     "PUSHER_APP_KEY",
	Secret:  "PUSHER_APP_SECRET",
	Cluster: "PUSHER_APP_CLUSTER",
	Secure:  true,
}

type Post struct {
	ID        int64  `json:"id"`
	Fullname  string `json:"fullname"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type PostCollection struct {
	Posts []Post `json:"items"`
}

func initialiseDatabase(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)

	if err != nil {
		panic(err)
	}

	if db == nil {
		panic("db nil")
	}

	return db
}

func migrateDatabase(db *sql.DB) {
	sql := `
            CREATE TABLE IF NOT EXISTS posts(
                    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
                    "fullname" TEXT,
                    "content" TEXT,
                    "timestamp" VARCHAR
            );
    `
	_, err := db.Exec(sql)
	if err != nil {
		panic(err)
	}
}

func getPosts(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		rows, err := db.Query("SELECT * FROM posts ORDER BY id DESC")
		if err != nil {
			panic(err)
		}

		fmt.Println(rows, err)

		defer rows.Close()

		result := PostCollection{}

		for rows.Next() {
			post := Post{}
			err2 := rows.Scan(&post.ID, &post.Fullname, &post.Content, &post.Timestamp)
			if err2 != nil {
				panic(err2)
			}

			result.Posts = append(result.Posts, post)
		}

		return c.JSON(http.StatusOK, result)
	}
}

func savePost(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		postFullname := c.FormValue("fullname")
		postContent := c.FormValue("content")

		location, _ := time.LoadLocation("Africa/Lagos")

		postTimestamp := time.Now().In(location).Format("02-03-2006 3PM")

		stmt, err := db.Prepare("INSERT INTO posts (fullname, content, timestamp) VALUES(?, ?, ?)")
		if err != nil {
			panic(err)
		}

		fmt.Println(postFullname, postContent)
		defer stmt.Close()

		result, err := stmt.Exec(postFullname, postContent, postTimestamp)
		if err != nil {
			panic(err)
		}

		insertedID, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}

		post := Post{
			ID:        insertedID,
			Fullname:  postFullname,
			Content:   postContent,
			Timestamp: postTimestamp,
		}
		client.Trigger("go-note", "notes", post)
		return c.JSON(http.StatusOK, post)
	}
}
