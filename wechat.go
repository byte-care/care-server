package main

import (
	"github.com/gin-gonic/gin"
)

func wechatGet(c *gin.Context) {
	echostr := c.Query("echostr")

	c.String(200, echostr)
}
