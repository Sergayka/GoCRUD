package main

import (
	"GoCRUD/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	controllers.InitDataBase()

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	router.Static("/public", "./public")

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil) // StatusOk = Код 200
	})

	router.GET("/persons", controllers.ReadPerson)
	router.POST("/persons", controllers.CreatePerson)

	router.GET("/persons/:id", controllers.GetPersonByID)
	router.PUT("/persons/:id", controllers.UpdatePerson)

	router.DELETE("/persons/:id", controllers.DeletePerson)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
