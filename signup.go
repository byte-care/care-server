package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	core "github.com/byte-care/care-core"
	"github.com/byte-care/care-server-core/model"
)

func signup(c *gin.Context) {
	email, ok := c.GetPostForm("email")
	if !ok {
		c.String(403, "No Email")
		return
	}

	password, ok := c.GetPostForm("password")
	if !ok {
		c.String(403, "No Password")
		return
	}

	code, ok := c.GetPostForm("code")
	if !ok {
		c.String(403, "No Code")
		return
	}

	codeHash, ok := c.GetPostForm("code_hash")
	if !ok {
		c.String(403, "No Code Hash")
		return
	}

	channelID, ok := c.GetPostForm("channel_id")
	if !ok {
		c.String(403, "No ChannelID")
		return
	}

	if channelID != email {
		c.String(403, "ChannelID not equal to Email")
		return
	}

	expectedCodeHash := core.HashString(code, secretKeyGlobal)

	if expectedCodeHash != codeHash {
		c.String(403, "Code is wrong")
		return
	}

	accessKey, err := generateKey()
	if err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	secretKey, err := generateKey()
	if err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	user := &model.User{
		Email:     email,
		Password:  hashPassword(password),
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	if err := db.Create(user).Error; err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	cEmail := &model.ChannelEmail{
		UserID:  user.ID,
		Address: email,
		Count:   100,
	}

	if err := db.Create(cEmail).Error; err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	token, err := generateIDToken(fmt.Sprint(user.ID))
	if err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}
