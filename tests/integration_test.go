package tests

import (
	"context"
	"encoding/json"
	// "encoding/json"
	"employee-service/handlers"
	"employee-service/models"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gorilla/mux"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gotest.tools/assert"
)

var client *mongo.Client

func TestMain(m *testing.M) {
	ctx := context.Background()
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo:latest",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForListeningPort(nat.Port("27017/tcp")),
		},
		Started: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer mongoC.Terminate(ctx)

	ip, err := mongoC.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	port, err := mongoC.MappedPort(ctx, "27017")
	if err != nil {
		log.Fatal(err)
	}

	mongoURI := "mongodb://" + ip + ":" + port.Port()
	client, err = mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestRegisterEmployee(t *testing.T) {
	handler := handlers.NewEmployeeHandler(client)

	r := mux.NewRouter()
	r.HandleFunc("/register", handler.RegisterEmployee).Methods("POST")

	// Load test data from file
	file, err := os.Open("testdata/register.json")
	assert.NilError(t, err)
	defer file.Close()

	payload, err := ioutil.ReadAll(file)
	assert.NilError(t, err)

	// Perform initial registration
	req, err := http.NewRequest("POST", "/register", strings.NewReader(string(payload)))
	assert.NilError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)

	// Attempt to register with the same email (expect conflict error)
	reqConflictEmail, err := http.NewRequest("POST", "/register", strings.NewReader(string(payload)))
	assert.NilError(t, err)

	rrConflictEmail := httptest.NewRecorder()
	r.ServeHTTP(rrConflictEmail, reqConflictEmail)

	assert.Equal(t, rrConflictEmail.Code, http.StatusConflict)

	// Modify payload to simulate registration with existing phone number
	var employee models.Employee
	err = json.Unmarshal(payload, &employee)
	assert.NilError(t, err)
	employee.Email = "another.email@example.com" // Change email to avoid email conflict, keep phone number for conflict

	modifiedPayload, err := json.Marshal(employee)
	assert.NilError(t, err)

	// Perform registration with existing phone number (expect conflict error)
	reqConflictPhone, err := http.NewRequest("POST", "/register", strings.NewReader(string(modifiedPayload)))
	assert.NilError(t, err)

	rrConflictPhone := httptest.NewRecorder()
	r.ServeHTTP(rrConflictPhone, reqConflictPhone)

	assert.Equal(t, rrConflictPhone.Code, http.StatusConflict)
}

func TestEmployeeById(t *testing.T) {
	handler := handlers.NewEmployeeHandler(client)

	r := mux.NewRouter()
	r.HandleFunc("/employee/{id}", handler.EmployeeById).Methods("GET")

	// Insert test employee
	collection := client.Database("testdb").Collection("employees")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	employee := models.Employee{
		Name:           "Jane Doe",
		Email:          "jane.doe@example.com",
		Age:            25,
		WorkExperience: 3,
	}
	res, err := collection.InsertOne(ctx, employee)
	assert.NilError(t, err)

	id := res.InsertedID.(primitive.ObjectID).Hex()

	// Perform valid request
	reqValid, err := http.NewRequest("GET", "/employee/"+id, nil)
	assert.NilError(t, err)

	rrValid := httptest.NewRecorder()
	r.ServeHTTP(rrValid, reqValid)

	assert.Equal(t, rrValid.Code, http.StatusOK)

	// Perform request with invalid ID (expect bad request error)
	reqInvalidID, err := http.NewRequest("GET", "/employee/invalid-id", nil)
	assert.NilError(t, err)

	rrInvalidID := httptest.NewRecorder()
	r.ServeHTTP(rrInvalidID, reqInvalidID)

	assert.Equal(t, rrInvalidID.Code, http.StatusBadRequest)

	// Perform request with non-existent ID (expect not found error)
	reqNonExistentID, err := http.NewRequest("GET", "/employee/123456789012345678901234", nil)
	assert.NilError(t, err)

	rrNonExistentID := httptest.NewRecorder()
	r.ServeHTTP(rrNonExistentID, reqNonExistentID)

	assert.Equal(t, rrNonExistentID.Code, http.StatusNotFound)
}

func TestEmployees(t *testing.T) {
	handler := handlers.NewEmployeeHandler(client)

	r := mux.NewRouter()
	r.HandleFunc("/employees", handler.Employees).Methods("GET")

	// Test case 1: Successful retrieval of employees
	reqSuccess, err := http.NewRequest("GET", "/employees", nil)
	assert.NilError(t, err)

	rrSuccess := httptest.NewRecorder()
	r.ServeHTTP(rrSuccess, reqSuccess)

	assert.Equal(t, rrSuccess.Code, http.StatusOK)

	// Test case 2: Simulate failure (e.g., database disconnection)
	// Disconnect MongoDB client to simulate an internal server error
	client.Disconnect(context.Background())

	reqFailure, err := http.NewRequest("GET", "/employees", nil)
	assert.NilError(t, err)

	rrFailure := httptest.NewRecorder()
	r.ServeHTTP(rrFailure, reqFailure)

	assert.Equal(t, rrFailure.Code, http.StatusInternalServerError)
}
