package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Employee struct {
    ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name          string             `bson:"name" json:"name"`
    Email         string             `bson:"email" json:"email"`
    Age           int                `bson:"age" json:"age"`
    WorkExperience int               `bson:"work_experience" json:"work_experience"`
    PhoneNumber   string             `bson:"phone_number" json:"phone_number"`
}
