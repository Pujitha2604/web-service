package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "employee-service/models"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
   // "go.mongodb.org/mongo-driver/mongo/options"
    "time"
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
    result, err := collection.InsertOne(ctx, employee)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(result)
}

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
            http.Error(w, "Employee not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(employee)
}

func (h *EmployeeHandler) Employees(w http.ResponseWriter, r *http.Request) {
    collection := h.client.Database("testdb").Collection("employees")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)
    var employees []models.Employee
    if err = cursor.All(ctx, &employees); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(employees)
}
