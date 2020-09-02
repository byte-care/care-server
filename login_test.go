package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	dropAndMigrate()

	const email = "h@h.com"
	const password = "pwd123456"

	createUser(email, password)

	// ✅ Success
	data := url.Values{"email": {email}, "password": {password}}

	w := testReq("post", "/login", data, nil, "")

	// ---
	assert.Equal(t, 200, w.Code)
	//

	// ❌ Failure
	data = url.Values{"email": {email}, "password": {"badpw"}}

	w = testReq("post", "/login", data, nil, "")
	// ---
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, 0, len(w.Body.String()))
	//
}
