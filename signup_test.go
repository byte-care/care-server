package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	core "github.com/byte-care/care-core"
	"github.com/byte-care/care-server-core/model"
)

func TestSignup(t *testing.T) {
	dropAndMigrate()

	const email = "h@h.com"
	const password = "pwd123456"

	code, _, err := generateCode(email)
	if err != nil {
		panic(err)
	}

	// ✅ Success
	data := url.Values{
		"email":      {email},
		"password":   {password},
		"code":       {code},
		"code_hash":  {core.HashString(code, secretKeyGlobal)},
		"channel_id": {email},
	}

	w := testReq("post", "/signup", data, nil, "")
	// ---
	assert.Equal(t, 200, w.Code)
	var user model.User
	db.Take(&user)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, hashPassword(password), user.Password)
	//

	// ❌ Failure for missing field
	data = url.Values{
		"email":      {email},
		"code":       {code},
		"code_hash":  {core.HashString(code, secretKeyGlobal)},
		"channel_id": {email},
	}

	w = testReq("post", "/signup", data, nil, "")
	// ---
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, "No Password", w.Body.String())
	//

	// ❌ Failure for email not equal to channel_id
	data = url.Values{
		"email":      {email},
		"password":   {password},
		"code":       {code},
		"code_hash":  {core.HashString(code, secretKeyGlobal)},
		"channel_id": {"fake_email"},
	}

	w = testReq("post", "/signup", data, nil, "")
	// ---
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, "ChannelID not equal to Email", w.Body.String())
	//

	// ❌ Failure for wrong code
	data = url.Values{
		"email":      {email},
		"password":   {password},
		"code":       {"223567"},
		"code_hash":  {core.HashString(code, secretKeyGlobal)},
		"channel_id": {email},
	}

	w = testReq("post", "/signup", data, nil, "")
	// ---
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, "Code is wrong", w.Body.String())
	//
}
