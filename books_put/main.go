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

	e.PUT("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id") // get the id from the URL parameter

		var book BookStore
		if err := c.Bind(&book); err != nil { // bind the request body to the book struct
			return c.String(http.StatusBadRequest, "Invalid input")
		}

		if book.BookName == "" || book.BookAuthor == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "title and author are required"})
		}

		book.ID = id // ensure the ID is set to the one from the URL

		filter := bson.M{"id": id}     // filter to find the book by ID
		update := bson.M{"$set": book} // update the book with the new data

		result, err := coll.UpdateOne(context.TODO(), filter, update) // perform the update
		// Check if the update was successful
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update book")
		}

		if result.MatchedCount == 0 { // no book found with the given ID
			return c.String(http.StatusNotFound, "Book not found")
		}

		return c.JSON(http.StatusOK, book)
	})

	// We start the server and bind it to port 3030. For future references, this
	// is the application's port and not the external one. For this first exercise,
	// they could be the same if you use a Cloud Provider. If you use ngrok or similar,
	// they might differ.
	// In the submission website for this exercise, you will have to provide the internet-reachable
	// endpoint: http://<host>:<external-port>
	e.Logger.Fatal(e.Start(":3030"))
}
