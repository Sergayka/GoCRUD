package controllers

import (
	"GoCRUD/models"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"os"
	"path/filepath"
)

var minioClient *minio.Client

var collection *mongo.Collection

func InitDataBase() {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017"
	}

	clientOptions := options.Client().ApplyURI(mongoURL)
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

func InitMinio() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
}

func CreatePerson(c *gin.Context) { // получил по башке за context
	// Получаем файл аватарки из формы
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get avatar"})
		return
	}
	defer file.Close()

	// Генерируем уникальное имя для файла
	fileName := fmt.Sprintf("%s%s", primitive.NewObjectID().Hex(), filepath.Ext(header.Filename))

	bucketName := "avatars"
	location := "us-east-1"

	// Проверка - существует ли бакет, если нет - создаем
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: location})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Определение политики
	policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": "*",
                "Action": [
                    "s3:GetObject"
                ],
                "Resource": [
                    "arn:aws:s3:::avatars/*"
                ]
            }
        ]
    }`

	// Cтавим чертову политику баскета
	err = minioClient.SetBucketPolicy(context.Background(), "avatars", policy)
	if err != nil {
		panic(err)
	}

	// Сохраняем файл в MinIO
	_, err = minioClient.PutObject(context.Background(), bucketName, fileName, file, header.Size, minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем внешний хост MinIO
	minioHost := os.Getenv("MINIO_EXTERNAL_HOST")
	if minioHost == "" {
		minioHost = "localhost:9000" // замените на ваш внешний хост или доменное имя
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", os.Getenv("MINIO_EXTERNAL_HOST"), bucketName, fileName)
	fmt.Println("Uploaded file URL:", fileURL)
	fmt.Println("http://%s/%s/%s", minioHost, bucketName, fileName)

	// Создаем пользователя
	var person models.Person
	person.ID = primitive.NewObjectID()
	person.FirstName = c.PostForm("firstName")
	person.LastName = c.PostForm("lastName")
	person.City = c.PostForm("city")
	person.AvatarURL = fileURL

	_, err = collection.InsertOne(context.Background(), person)
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
