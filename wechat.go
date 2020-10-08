package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/byte-care/care-server-core/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/silenceper/wechat/message"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func wechatGet(c *gin.Context) {
	echostr := c.Query("echostr")

	c.String(200, echostr)
}

func wechatPost(c *gin.Context) {
	var req message.MixMessage
	if err := c.ShouldBindXML(&req); err != nil {
		c.String(403, err.Error())
		log.Println(err.Error())
		return
	}

	if req.MsgType == message.MsgTypeEvent {
		openID := req.FromUserName

		if req.Event == message.EventClick {
			if req.EventKey == "task_list" {
				// OpenID -> UserID
				var channelWechat model.ChannelWeChat
				result := db.Select("user_id").Where("mp_open_id = ?", openID).First(&channelWechat)
				if result.Error != nil {
					if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
						log.Println(result.Error)
						c.Status(403)
						return
					}
					resp := constructTextResp(`请在<a href="https://console.bytecare.xyz/">控制台</a>绑定微信`, req)
					c.XML(200, resp)
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
						var statusIcon string
						if task.status == 0 {
							statusIcon = "⏳"
						} else if task.status == 1 {
							statusIcon = "✔"
						} else if task.status == 2 {
							statusIcon = "❌"
						} else if task.status == 3 {
							statusIcon = "❌"
						}

						var briefTopic string
						if len(task.topic) < 10 {
							briefTopic = task.topic
						} else {
							briefTopic = fmt.Sprintf("%s...", task.topic[:10])
						}

						buf.WriteString(fmt.Sprintf(`%d %s %s\n<a href="https://www.baidu.com/">详情</a>\n\n`, task.id, statusIcon, briefTopic))
					}
					content = buf.String()
				}

				resp := constructTextResp(content, req)

				c.XML(200, resp)

				return
			}
		} else if req.Event == message.EventSubscribe || req.Event == message.EventScan {
			userIDString := req.EventKey[8:]
			userID, err := strconv.ParseUint(userIDString, 10, 64)
			if err != nil {
				log.Println(err.Error())
				c.String(403, "")
				return
			}

			openIDString := string(openID)

			var channelWechat model.ChannelWeChat
			result := db.Select("mp_open_id").Where("user_id = ?", userID).First(&channelWechat)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					log.Println(result.Error)
					return
				}

				newChannelWechat := &model.ChannelWeChat{
					UserID:   uint(userID),
					MPOpenID: openIDString,
				}

				if err := db.Create(newChannelWechat).Error; err != nil {
					log.Println(err.Error())
					c.String(403, "")
					return
				}

				c.String(202, "")
				return
			}

			if channelWechat.MPOpenID != openIDString {
				db.Model(&channelWechat).Update("mp_open_id", openIDString)
			}

			c.Status(202)
			return
		}
	}

	resp := constructTextResp(`ByteCare`, req)

	c.XML(200, resp)
}

type wechatQRRespStruct struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds uint32 `json:"expire_seconds"`
}

func weChatQR(c *gin.Context) {
	userID, ok := c.GetPostForm("id")
	if !ok {
		c.String(403, "No ID")
		return
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", mpAccessToken)

	bodyString := fmt.Sprintf(`{"expire_seconds": 1592000, "action_name": "QR_SCENE", "action_info": {"scene": {"scene_id": %s}}}`, userID)
	resp, err := http.Post(url, "application/json", strings.NewReader(bodyString))
	if err != nil {
		log.Println(err)
		c.String(403, "")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		c.String(403, "")
		return
	}

	var wechatResp wechatQRRespStruct
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		log.Println(err)
		c.String(403, "")
		return
	}

	if wechatResp.ExpireSeconds == 0 {
		log.Println(body)
		c.String(403, "")
		return
	}

	c.String(202, wechatResp.Ticket)
	return
}

func constructTextResp(content string, req message.MixMessage) *message.Text {
	resp := message.NewText(content)
	resp.SetToUserName(req.FromUserName)
	resp.SetFromUserName(req.ToUserName)
	resp.SetCreateTime(req.CreateTime)
	resp.SetMsgType("text")
	return resp
}
