/*
	Note: To get the user details, we need to pass user id as "_id" param with the url
	Note: To get the contact details of a User, we need to pass user id as _id param with the url
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type User struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	UserID string `json:"userId" bson:"userId" binding: "reqired"`
	Name string   `json:"name," bson:"name"`
	PhoneNumber string `json:"phoneNumber" bson:"phoneNumber"`
	Email string   `json:"email" bson:"email"`
	TimeStamp string `json:"timeStamp" bson: "timeStamp"`
	DateOfBirth string `json:"dob,omitempty" bson: "dob,omitempty"`
}

type Contact struct{
	ContactID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
    UserOneId string `json:"user_id_one,omitempty" bson:"user_id_one,omitempty"`
    UserTwoId string `json:"user_id_two,omitempty" bson:"user_id_two,omitempty"`
    TimeOfContact string `json:"toc,omitempty" bson:"toc,omitempty"`
}


func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	id := request.URL.Query()["_id"]                                   // Fetching param "_id" from url

	collection := client.Database("appointy").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var userTemp User
		cursor.Decode(&userTemp)
		fmt.Println(userTemp.ID.String()[10:34], id[0])                 // Fetching user with same id as Param "_id"
		if userTemp.ID.String()[10:34] == id[0]{
			json.NewEncoder(response).Encode(userTemp)
			break
		}
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	
}

func GetContactEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	//var user User
	id := request.URL.Query()["_id"][0]
	var c_ids []string                      // Stores IDs of all contacts of user with "_id" which is passed as param
	var users []User                        // Stores all Users in contacts whose IDs are in c_ids
	
	c_collection := client.Database("appointy").Collection("contacts")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	c_cursor, _ := c_collection.Find(ctx, bson.M{})
	
	for c_cursor.Next(ctx) {
		var contactTemp Contact
		c_cursor.Decode(&contactTemp)
		if contactTemp.UserOneId == id{
			c_ids = append(c_ids, contactTemp.UserTwoId)
		}
		if contactTemp.UserTwoId == id{
			c_ids = append(c_ids, contactTemp.UserOneId)
		}
	}
//	fmt.Println("c_ids",c_ids) To check whether c_ids is getting desired values
	if c_err := c_cursor.Err(); c_err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + c_err.Error() + `" }`))
		return
	}
	u_collection := client.Database("appointy").Collection("users")
	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	u_cursor, _ := u_collection.Find(ctx, bson.M{})
	for _,c_id := range c_ids{                        // Getting Contact's ID as c_id from c_ids
		for u_cursor.Next(ctx) {                      // Storing Users with selected IDs
			var userTemp User
			u_cursor.Decode(&userTemp)
			if userTemp.ID.String()[10:34] == c_id{
				users = append(users, userTemp)
			}
		}
		if u_err := u_cursor.Err(); u_err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + u_err.Error() + `" }`))
			return
		}
	}
	if c_err := c_cursor.Err(); c_err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + c_err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(users)
	defer u_cursor.Close(ctx)
	defer c_cursor.Close(ctx)
}



func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var user User
	_ = json.NewDecoder(request.Body).Decode(&user)
	collection := client.Database("appointy").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	user.TimeStamp = time.Now().String()
	result, _ := collection.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(user)
	fmt.Println(result)
}	

func CreateContactEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var contact Contact
	_ = json.NewDecoder(request.Body).Decode(&contact)
	collection := client.Database("appointy").Collection("contacts")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	contact.TimeOfContact = time.Now().String()
	result, _ := collection.InsertOne(ctx, contact)
	json.NewEncoder(response).Encode(contact)
	fmt.Println(result)
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://deathcreed:appointy@cluster0.tta7y.mongodb.net/test")
	client, _ = mongo.Connect(ctx,clientOptions)
	http.HandleFunc("/user", CreateUserEndpoint) // to create a new user
	http.HandleFunc("/userID", GetUserEndpoint) // to get info of a user
	http.HandleFunc("/contact", CreateContactEndpoint) // to create a contact
	http.HandleFunc("/contacts",GetContactEndpoint) // to get contacts of a user
	log.Fatal(http.ListenAndServe(":12345", nil))
}