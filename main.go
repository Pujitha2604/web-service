package main

import (
	"context"
	"employee-service/db"
	"employee-service/handlers"
	"employee-service/routes"
	"log"
	"net/http"
	"time"
)

func main() {
	mongoURI := "mongodb+srv://task3-shreeraj:YIXZaFDnEmHXC3PS@cluster0.0elhpdy.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	if mongoURI == "" {
		log.Fatal("MONGO_URI environment variable is required")
	}

	client, err := db.ConnectMongoDB(mongoURI, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DisconnectMongoDB(client, context.Background())

	handler := handlers.NewEmployeeHandler(client)
	r := routes.RegisterRoutes(handler)

	http.Handle("/", r)
	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
