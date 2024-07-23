package handlers

import (
	"context"
	"employee-service/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h *EmployeeHandler) EmployeeById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	collection := h.client.Database("testdb").Collection("employees")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var employee models.Employee
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Employee with this ID doesn't exist", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(employee)
}
