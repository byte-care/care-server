package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/byte-care/care-server-core/model"
)

func sendEmail(c *gin.Context) {
	specificParameter := make(map[string]string)

	msg, ok := c.GetPostForm("msg")
	if !ok {
		c.String(403, "No msg")
		return
	}
	specificParameter["msg"] = msg

	topic, ok := c.GetPostForm("topic")
	if !ok {
		c.String(403, "No topic")
		return
	}
	specificParameter["topic"] = topic

	// Check Signature and Get User(Contain id, secret_key)
	user, err := checkSignature(c, specificParameter)
	if err != nil {
		c.String(403, err.Error())
		return
	}

	var cEmail model.ChannelEmail
	query := db.Model(&user).Related(&cEmail)
	if err := query.Error; err != nil {
		c.String(403, "Doesn't Find Email Belong to the User")
		return
	}

	rowsAffected := db.Model(&cEmail).Where("count > 0").UpdateColumn("count", gorm.Expr("count - ?", 1)).RowsAffected
	if rowsAffected == 0 {
		c.String(403, "Email Count not Enough")
		return
	}

	err = serviceGlobal.email(cEmail.Address, topic, msg)
	if err != nil {
		log.Println(err.Error())
		c.String(403, "Fail to Send Email")
		return
	}

	c.JSON(200, gin.H{})
}
