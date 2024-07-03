package controllers

import (
	"GoCRUD/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

var collection *mongo.Collection

func InitDataBase() {
	URI := "mongodb://localhost:27017"

	clientOptions := options.Client().ApplyURI(URI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	collection = client.Database("CRUD").Collection("test")
}

func CreatePerson(c *gin.Context) { // получил по башке за context
	var person models.Person

	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	person.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(context.Background(), person)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, person)
}

func ReadPerson(c *gin.Context) {
	var persons []models.Person
	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, context.Background())

	for cursor.Next(context.Background()) {
		var person models.Person
		err := cursor.Decode(&person)
		if err != nil {
			return
		}
		persons = append(persons, person)
	}

	c.JSON(http.StatusOK, persons)
}

func UpdatePerson(c *gin.Context) {
	id := c.Param("id")

	var person models.Person
	if err := c.ShouldBindJSON(&person); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"first_name": person.FirstName,
		"last_name":  person.LastName,
		"city":       person.City,
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, bson.M{"$set": update})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, person)
}

func DeletePerson(c *gin.Context) {
	id := c.Param("id")
	person, _ := primitive.ObjectIDFromHex(id)
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": person})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, person)
}

func GetPersonByID(c *gin.Context) {
	id := c.Param("id")

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var person models.Person
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&person)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, person)
}
