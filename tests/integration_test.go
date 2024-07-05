package tests

import (
    "context"
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

    file, err := os.Open("testdata/register.json")
    assert.NilError(t, err)
    defer file.Close()

    payload, err := ioutil.ReadAll(file)
    assert.NilError(t, err)

    req, err := http.NewRequest("POST", "/register", strings.NewReader(string(payload)))
    assert.NilError(t, err)

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, rr.Code, http.StatusOK)
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
        Name:          "Jane Doe",
        Email:         "jane.doe@example.com",
        Age:           25,
        WorkExperience: 3,
    }
    res, err := collection.InsertOne(ctx, employee)
    assert.NilError(t, err)

    id := res.InsertedID.(primitive.ObjectID).Hex()

    // Replace placeholder in test file with actual ID
    inputFile := "testdata/employee_by_id.json"
    fileData, err := ioutil.ReadFile(inputFile)
    assert.NilError(t, err)
    fileData = []byte(strings.Replace(string(fileData), "PLACEHOLDER_FOR_ID", id, 1))

    req, err := http.NewRequest("GET", "/employee/"+id, strings.NewReader(string(fileData)))
    assert.NilError(t, err)

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, rr.Code, http.StatusOK)
}

func TestEmployees(t *testing.T) {
    handler := handlers.NewEmployeeHandler(client)

    r := mux.NewRouter()
    r.HandleFunc("/employees", handler.Employees).Methods("GET")

    req, err := http.NewRequest("GET", "/employees", nil)
    assert.NilError(t, err)

    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    assert.Equal(t, rr.Code, http.StatusOK)
}
