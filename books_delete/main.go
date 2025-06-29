package main

import (
	"context"
	"net/http"
	"time"

	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

	e.DELETE("/api/books/:id", func(c echo.Context) error {
		id := c.Param("id")                                   // get the id from the URL parameter
		filter := bson.M{"id": id}                            // filter to find the book by ID
		result, err := coll.DeleteOne(context.TODO(), filter) // perform the delete
		// Check if the delete was successful
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to delete book")
		}
		if result.DeletedCount == 0 { // no book found with the given ID
			return c.String(http.StatusNotFound, "Book not found")
		}

		// return status 200 success
		return c.NoContent(http.StatusOK)
	})

	// We start the server and bind it to port 3030. For future references, this
	// is the application's port and not the external one. For this first exercise,
	// they could be the same if you use a Cloud Provider. If you use ngrok or similar,
	// they might differ.
	// In the submission website for this exercise, you will have to provide the internet-reachable
	// endpoint: http://<host>:<external-port>
	e.Logger.Fatal(e.Start(":3030"))
}
