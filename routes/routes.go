package routes

import (
    "employee-service/handlers"
    "github.com/gorilla/mux"
    "net/http"
)

func RegisterRoutes(handler *handlers.EmployeeHandler) *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/register", handler.RegisterEmployee).Methods(http.MethodPost)
    r.HandleFunc("/employee/{id}", handler.EmployeeById).Methods(http.MethodGet)
    r.HandleFunc("/employees", handler.Employees).Methods(http.MethodGet)
    return r
}
