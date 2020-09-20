package main

import (
	"errors"
	"github.com/byte-care/care-server-core/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"log"
)

func setDefaultChannel(c *gin.Context) {
	id, ok := c.GetPostForm("id")
	if !ok {
		c.String(403, "No ID")
		return
	}

	channel, ok := c.GetPostForm("channel")
	if !ok {
		c.String(403, "No Channel")
		return
	}

	if channel == "0" {
		db.Model(&model.User{}).Where("id = ?", id).Update("default_channel", 0)
		c.Status(202)
		return
	} else if channel == "1" {
		var cWechat model.ChannelWeChat
		result := db.Select("id").Where("user_id = ?", id).First(&cWechat)
		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Println(result.Error)
				c.Status(403)
				return
			}
			c.String(403, "unsubscribe")
			return
		}
		db.Model(&model.User{}).Where("id = ?", id).Update("default_channel", 1)
		c.Status(202)
		return
	}

	c.Status(403)
	return
}
