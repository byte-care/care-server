package main

import "github.com/gin-gonic/gin"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/send-email-code", sendEmailCode)
	r.GET("/oauth/github", OAuthGitHub)
	r.POST("/signup", signup)
	r.POST("/login", login)
	r.POST("/view-key", viewKey)

	r.GET("/wechat", wechatGet)
	r.POST("/wechat", wechatPost)

	r.POST("/send-email", sendEmail)

	r.GET("/log/pub", logPub)

	return r
}

func main() {
	setup(false)
	serviceGlobal = realService{}
	wechatNotifyServiceGlobal = realWechatNotifyService{}

	r := setupRouter()
	_ = r.Run(":8080")
}
