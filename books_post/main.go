package main

import (
	"context"
	"net/http"
	"time"

	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
// More on these "tags" like `bson:"_id,omitempty"`: https://go.dev/wiki/Well-known-struct-tags
type BookStore struct {
	MongoID     primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	ID          string             `bson:"id" json:"id"`
	BookName    string             `bson:"title" json:"title"`
	BookAuthor  string             `bson:"author" json:"author"`
	BookEdition string             `bson:"edition,omitempty" json:"edition"`
	BookPages   string             `bson:"pages,omitempty" json:"pages"`
	BookYear    string             `bson:"year,omitempty" json:"year"`
}

func main() {
	// Connect to the database. Such defer keywords are used once the local
	// context returns; for this case, the local context is the main function
	// By user defer function, we make sure we don't leave connections
	// dangling despite the program crashing. Isn't this nice? :D
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TODO: make sure to pass the proper username, password, and port
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb:testmongo@localhost:27017"))
	databaseUri := os.Getenv("DATABASE_URI")
	if databaseUri == "" {
		databaseUri = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseUri))

	// This is another way to specify the call of a function. You can define inline
	// functions (or anonymous functions, similar to the behavior in Python)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// connection to the database and collection
	coll := client.Database("exercise-3").Collection("information")

	// Here we prepare the server
	e := echo.New()

	// Log the requests. Please have a look at echo's documentation on more
	// middleware
	e.Use(middleware.Logger())

	e.POST("/api/books", func(c echo.Context) error {
		var book BookStore
		if err := c.Bind(&book); err != nil { // bind the request body to the book struct
			return c.String(http.StatusBadRequest, "Invalid input")
		}

		if book.ID == "" || book.BookName == "" || book.BookAuthor == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "id, title and author are required"})
		}

		// Check if the book already exists
		filter := bson.M{"id": book.ID}
		count, err := coll.CountDocuments(context.TODO(), filter)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to check for duplicates")
		}
		if count > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Book with the same ID already exists"})
		}

		// Insert the new book
		result, err := coll.InsertOne(context.TODO(), book)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to insert book")
		}

		book.MongoID = result.InsertedID.(primitive.ObjectID)
		return c.JSON(http.StatusCreated, book)
	})

	// We start the server and bind it to port 3030. For future references, this
	// is the application's port and not the external one. For this first exercise,
	// they could be the same if you use a Cloud Provider. If you use ngrok or similar,
	// they might differ.
	// In the submission website for this exercise, you will have to provide the internet-reachable
	// endpoint: http://<host>:<external-port>
	e.Logger.Fatal(e.Start(":3030"))
}
