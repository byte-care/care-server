package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/byte-care/care-server-core/model"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"io/ioutil"
	"log"
	"net/http"
)

type GitHubProfile struct {
	ID    uint
	Email string
}

func OAuthGitHub(c *gin.Context) {
	conf := &oauth2.Config{
		ClientID:     GitHubClientID,
		ClientSecret: GitHubClientSecret,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	ctx := context.Background()

	token, err := conf.Exchange(ctx, c.Query("code"))
	if err != nil {
		c.String(403, err.Error())
		log.Println(err.Error())

		return
	}

	client := conf.Client(ctx, token)
	response, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.String(403, err.Error())
		log.Println(err.Error())

		return
	}

	defer response.Body.Close()
	profileBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.String(403, err.Error())
		log.Println(err.Error())

		return
	}

	var profile GitHubProfile

	err = json.Unmarshal(profileBytes, &profile)
	if err != nil {
		c.String(403, err.Error())
		log.Println(err.Error())

		return
	}

	var oauthInfo model.OAuthGitHub
	result := db.Select("user_id").Where("git_hub_id = ?", profile.ID).First(&oauthInfo)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
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
			Email:     profile.Email,
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
			Address: profile.Email,
			Count:   100,
		}

		if err := db.Create(cEmail).Error; err != nil {
			log.Println(err.Error())
			c.String(403, "")
			return
		}

		oGitHub := &model.OAuthGitHub{
			UserID:   user.ID,
			GitHubID: profile.ID,
		}

		if err := db.Create(oGitHub).Error; err != nil {
			log.Println(err.Error())
			c.String(403, "")
			return
		}
	}

	jwtToken, err := generateIDToken(fmt.Sprint(oauthInfo.UserID))
	if err != nil {
		log.Println(err.Error())
		c.String(403, "")
		return
	}

	c.SetCookie("token", jwtToken, 72000, "/", ".bytecare.xyz", true, false)

	c.Redirect(http.StatusFound, "https://www.bytecare.xyz/account-general.html")
}
