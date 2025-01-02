package main

import (
	"encoding/json"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

var books = []book{}

const booksFile = "books.json"

func init() {
	loadBooksFromFile()
}

func loadBooksFromFile() {
	data, _ := ioutil.ReadFile(booksFile)
	err := json.Unmarshal(data, &books)
	if err != nil {
		return
	}
}

func saveBooksToFile() {
	data, _ := json.MarshalIndent(books, "", " ")
	ioutil.WriteFile(booksFile, data, 0644)
}

func getBooks(c *gin.Context) {
	title := c.Query("titie")
	author := c.Query("author")

	filteredBooks := books

	if title != "" {
		filteredBooks = []book{}
		for _, book := range books {
			if book.Title == title {
				filteredBooks = append(filteredBooks, book)
			}
		}
	}

	if author != "" {
		tempBooks := []book{}
		for _, book := range filteredBooks {
			if book.Author == author {
				tempBooks = append(tempBooks, book)
			}
		}
		filteredBooks = tempBooks
	}
	c.IndentedJSON(http.StatusOK, filteredBooks)
}

func getBooksById(id string) (*book, error) {
	for i, book := range books {
		if book.ID == id {
			return &books[i], nil
		}
	}
	return nil, errors.New("Book not found!")
}

func bookById(c *gin.Context) {
	id := c.Param("id")
	book, err := getBooksById(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"Message": "Book not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, book)
}

func createBook(c *gin.Context) {
	var newBook book

	if err := c.BindJSON(&newBook); err != nil {
		return
	}
	books = append(books, newBook)
	saveBooksToFile()
	c.IndentedJSON(http.StatusCreated, newBook)
}

func checkoutBook(c *gin.Context) {
	id, ok := c.GetQuery("id")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Message": "Missing id query parameter"})
		return
	}

	book, err := getBooksById(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"Message": "Book not found"})
		return
	}

	if book.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Message": "Book not available"})
		return
	}
	book.Quantity -= 1
	saveBooksToFile()
	c.IndentedJSON(http.StatusOK, book)
}

func returnBook(c *gin.Context) {
	id, ok := c.GetQuery("id")

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"Message": "Missing id query parameter"})
		return
	}

	book, err := getBooksById(id)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"Message": "Book not found"})
		return
	}
	book.Quantity += 1
	saveBooksToFile()
	c.IndentedJSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")
	for i, book := range books {
		if book.ID == id {
			books = append(books[:i], books[i+1:]...)
			saveBooksToFile()
			c.IndentedJSON(http.StatusOK, gin.H{"Message": "Book deleted successfully"})
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"Message": "Book not found"})
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	router.GET("/books", getBooks)
	router.GET("/books/:id", bookById)
	router.POST("/books", createBook)
	router.DELETE("/books/:id", deleteBook)
	router.PATCH("/checkout", checkoutBook)
	router.PATCH("/return", returnBook)

	router.Run("localhost:8080")
}
