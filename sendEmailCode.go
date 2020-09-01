package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func sendEmailCode(c *gin.Context) {
	email, ok := c.GetPostForm("email")
	if !ok {
		c.String(403, "No Email")
		return
	}

	raw, token, err := generateCode(email)
	if err != nil {
		c.String(403, "Fail to Generate Code")
		return
	}

	err = serviceGlobal.email(email, "验证码", raw)
	if err != nil {
		log.Println(err)
		c.String(403, "Fail to Send Email Code")
		return
	}

	c.JSON(200, gin.H{
		"token": token,
	})
}
