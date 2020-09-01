package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"

	core "github.com/byte-care/care-core"
)

var router *gin.Engine

func testReq(method string, urlString string, data interface{}, commonParameter *core.CommonParameter, signature string) *httptest.ResponseRecorder {
	var body io.Reader

	if method == "post" {
		switch v := reflect.ValueOf(data); v.Kind() {
		case reflect.String:
			body = bytes.NewBuffer([]byte(v.String()))
		default:
			formData := data.(url.Values)
			body = strings.NewReader(formData.Encode())
		}
	}

	var req *http.Request
	if method == "post" {
		req, _ = http.NewRequest("POST", urlString, body)
	} else {
		req, _ = http.NewRequest("GET", urlString, nil)
	}

	switch v := reflect.ValueOf(data); v.Kind() {
	case reflect.String:
		req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	default:
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if commonParameter != nil {
		req.Header.Set("Care-Key", commonParameter.AccessKey)
		req.Header.Set("Care-Timestamp", commonParameter.Timestamp)
		req.Header.Set("Care-Nonce", commonParameter.SignatureNonce)

		req.Header.Set("Care-Signature", signature)
	}

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	return w
}

func createUser(email string, password string) *httptest.ResponseRecorder {
	raw, _, err := generateCode(email)
	if err != nil {
		panic(err)
	}

	data := url.Values{
		"email":      {email},
		"password":   {password},
		"code":       {raw},
		"code_hash":  {core.HashString(raw, secretKeyGlobal)},
		"channel_id": {email},
	}

	w := testReq("post", "/signup", data, nil, "")

	if w.Code != 200 {
		panic("createUser Fail")
	}

	return w
}
