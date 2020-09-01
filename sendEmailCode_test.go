package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendEmailCode(t *testing.T) {
	// ✅ Success
	data := url.Values{"email": {"h.tsai@hotmail.com"}}
	w := testReq("post", "/send-email-code", data, nil, "")

	// ---
	assert.Equal(t, 200, w.Code)
	//

	// ❌ Failure
	data = url.Values{}
	w = testReq("post", "/send-email-code", data, nil, "")

	// ---
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, "No Email", w.Body.String())
	//
}
