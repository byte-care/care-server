package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"log"

	"github.com/byte-care/care-server-core/model"
)

func viewKey(c *gin.Context) {
	id, ok := c.GetPostForm("id")
	if !ok {
		c.String(403, "No ID")
		return
	}

	var user model.User
	result := db.Select("id, access_key, secret_key").Where("id = ?", id).First(&user)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
		}
		c.String(403, "Not Found User")
		return
	}

	c.JSON(200, gin.H{
		"AccessKey": user.AccessKey,
		"SecretKey": user.SecretKey,
	})
}
