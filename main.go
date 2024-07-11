package main

import (
	"context"
	"employee-service/handlers"
	"employee-service/routes"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoURI := "mongodb+srv://task3-shreeraj:YIXZaFDnEmHXC3PS@cluster0.0elhpdy.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	handler := handlers.NewEmployeeHandler(client)

	r := routes.RegisterRoutes(handler)

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
