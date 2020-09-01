package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/send-email-code", sendEmailCode)
	r.POST("/signup", signup)

	return r
}

func main() {
	setup(false)
	serviceGlobal = realService{}

	r := setupRouter()
	r.Run(":8080")
}
