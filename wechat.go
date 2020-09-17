package main

import (
	"github.com/gin-gonic/gin"
	"github.com/silenceper/wechat/message"
)

func wechatGet(c *gin.Context) {
	echostr := c.Query("echostr")

	c.String(200, echostr)
}

func wechatPost(c *gin.Context) {
	var req message.MixMessage
	if err := c.ShouldBindXML(&req); err != nil {
		c.String(403, err.Error())
		return
	}

	resp := message.NewText(`ByteCare`)
	resp.SetToUserName(req.FromUserName)
	resp.SetFromUserName(req.ToUserName)
	resp.SetCreateTime(req.CreateTime)
	resp.SetMsgType("text")

	c.XML(200, resp)
}
