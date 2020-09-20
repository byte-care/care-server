package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/byte-care/care-server-core/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/silenceper/wechat/message"
	"log"
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

	if req.MsgType == message.MsgTypeEvent {
		if req.Event == message.EventClick {
			if req.EventKey == "task_list" {
				// OpenID -> UserID
				openID := req.FromUserName
				var channelWechat model.ChannelWeChat
				result := db.Select("user_id").Where("mp_open_id = ?", openID).First(&channelWechat)
				if result.Error != nil {
					if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
						log.Println(result.Error)
					}
					c.String(403, "Not Found User")
					return
				}
				userId := channelWechat.UserID

				// get brief task list
				taskList, err := serviceGlobal.getBriefTaskList(fmt.Sprint(userId))
				if err != nil {
					log.Println(err.Error())
					return
				}

				var content string

				if len(taskList) == 0 {
					content = "Empty Task List"
				} else {
					buf := bytes.Buffer{}
					for _, task := range taskList {
						buf.WriteString(fmt.Sprintf("%d %s\n", task.status, task.topic))
					}
					content = buf.String()
				}

				resp := constructTextResp(content, req)

				c.XML(200, resp)

				return
			}
		}
	}

	resp := constructTextResp(`ByteCare`, req)

	c.XML(200, resp)
}

func constructTextResp(content string, req message.MixMessage) *message.Text {
	resp := message.NewText(content)
	resp.SetToUserName(req.FromUserName)
	resp.SetFromUserName(req.ToUserName)
	resp.SetCreateTime(req.CreateTime)
	resp.SetMsgType("text")
	return resp
}
