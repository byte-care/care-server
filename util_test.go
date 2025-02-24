package main

import (
	"github.com/byte-care/care-server-core/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	setup(true)
	router = setupRouter()
	serviceGlobal = mockService{}
	wechatNotifyServiceGlobal = mockWechatNotifyService{}
	emailNotifyServiceGlobal = mockEmailNotifyService{}
}

func dropDB() {
	db.DropTable(&model.User{})
	db.DropTable(&model.ChannelEmail{})
	db.DropTable(&model.ChannelWeChat{})
}

func dropAndMigrate() {
	dropDB()
	autoMigrate()
}

func TestGenerateCode(t *testing.T) {
	r, _, err := generateCode("main@kan.fun")
	assert.Equal(t, nil, err)
	assert.Equal(t, 6, len(r))
}
