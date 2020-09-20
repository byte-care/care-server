package main

import (
	"github.com/gin-gonic/gin"
	"time"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/oauth/github", OAuthGitHub)
	r.POST("/wechat-qr", weChatQR)

	r.POST("/send-email-code", sendEmailCode)
	r.POST("/signup", signup)
	r.POST("/login", login)
	r.POST("/view-key", viewKey)
	r.POST("/set-default-channel", setDefaultChannel)

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

	setMPAccessToken()
	ticker := time.NewTicker(100 * time.Minute)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			setMPAccessToken()
		}
	}()

	r := setupRouter()
	_ = r.Run(":8080")
}
