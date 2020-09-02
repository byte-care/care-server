package main

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/byte-care/care-server-core/model"
)

func login(c *gin.Context) {
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

	var user model.User
	result := db.Select("id").Where("email = ? AND password = ?", email, hashPassword(password)).First(&user)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
		}
		c.String(403, "")
		return
	}

	token, err := generateIDToken(fmt.Sprint(user.ID))
	if err != nil {
		log.Println(err)
		c.String(403, "")
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}
