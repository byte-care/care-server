package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func bin(c *gin.Context) {
	platform := c.Query("platform")
	if platform == "" {
		c.String(403, "No Platform")
		return
	}

	if (platform != "linux") && (platform != "windows") {
		c.String(403, "Platform not Supported")
		return
	}

	result, err := serviceGlobal.bin(platform)
	if err != nil {
		log.Println(err)
		c.Status(502)
		return
	}

	c.String(200, result)
	return
}
