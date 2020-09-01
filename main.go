package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/signup", signup)

	return r
}

func main() {
	setup(false)
	serviceGlobal = realService{}

	r := setupRouter()
	r.Run(":8080")
}
