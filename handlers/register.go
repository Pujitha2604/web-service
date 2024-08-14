package handlers

import (
	"context"
	"employee-service/models"
	"encoding/json"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmployeeHandler struct {
	client *mongo.Client
}

func NewEmployeeHandler(client *mongo.Client) *EmployeeHandler {
	return &EmployeeHandler{client: client}
}

func (h *EmployeeHandler) RegisterEmployee(w http.ResponseWriter, r *http.Request) {
	var employee models.Employee
	if err := json.NewDecoder(r.Body).Decode(&employee); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	collection := h.client.Database("testdb").Collection("employees")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if employee with the same email already exists
	err := collection.FindOne(ctx, bson.M{"email": employee.Email}).Err()
	if err != mongo.ErrNoDocuments {
		if err == nil {
			http.Error(w, "Employee with this email already exists", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if employee with the same phone number already exists
	err = collection.FindOne(ctx, bson.M{"phone_number": employee.PhoneNumber}).Err()
	if err != mongo.ErrNoDocuments {
		if err == nil {
			http.Error(w, "Employee with this phone number already exists", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Insert new employee
	employee.ID = primitive.NewObjectID() // Ensure the employee ID is set
	_, err = collection.InsertOne(ctx, employee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // Return HTTP 200 OK upon successful registration
	w.Write([]byte("Employee is Registered\n"))
}
