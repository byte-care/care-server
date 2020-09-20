package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"

	core "github.com/byte-care/care-core"
)

type notifyService interface {
	//logPubResult(userID uint, topic string, isSuccessful bool) error
}

type realWechatNotifyService struct {
}

type mockWechatNotifyService struct {
}

type service interface {
	email(address string, subject string, body string) error
	sms(number string, code string) error
	bin(platform string) ([]string, error)
	getBriefTaskList(userId string) (result []task, err error)
	logGetToEnd(reversedID string, fromHead bool, lastAutoID int64) ([]string, int64, error)
	newTask(reversedUserID string, topic string, _type int) (taskID int64, err error)
	updateTaskStatus(reversedUserID string, taskID int64, status int) (err error)
	newLog(taskID int64, content string) error
	weChatGetAccessToken() (string, error)
}

type realService struct {
}

type mockService struct {
}

func (s realService) email(address string, subject string, body string) error {
	request := requests.NewCommonRequest()
	request.Domain = "dm.aliyuncs.com"
	request.Version = "2015-11-23"
	request.ApiName = "SingleSendMail"

	request.QueryParams["AccountName"] = "no-reply@bytecare.xyz"
	request.QueryParams["AddressType"] = "1"
	request.QueryParams["ReplyToAddress"] = "false"
	request.QueryParams["ToAddress"] = address
	request.QueryParams["Subject"] = subject

	if core.IsAllWhiteChar(body) {
		request.QueryParams["HtmlBody"] = "<html></html>"
	} else {
		request.QueryParams["TextBody"] = body
	}

	_, err := clientGlobal.ProcessCommonRequest(request)
	if err != nil {
		return err
	}

	return nil
}

//goland:noinspection GoUnusedParameter
func (s mockService) email(address string, subject string, body string) error {
	return nil
}

func (s realService) sms(number string, code string) error {
	request := requests.NewCommonRequest()
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"

	request.QueryParams["PhoneNumbers"] = number
	request.QueryParams["SignName"] = "Progress"
	request.QueryParams["TemplateCode"] = "SMS_185811363"
	request.QueryParams["TemplateParam"] = fmt.Sprintf("{\"code\":\"%s\"}", code)

	_, err := clientGlobal.ProcessCommonRequest(request)
	if err != nil {
		return err
	}

	return nil
}

//goland:noinspection GoUnusedParameter
func (s mockService) sms(number string, code string) error {
	return nil
}

func (s realService) bin(platform string) (result []string, err error) {
	bucketName := "kan-bin"

	bucket, err := ossClientGlobal.Bucket(bucketName)
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/", platform)

	marker := ""
	for {
		lsRes, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(path))
		if err != nil {
			return nil, err
		}

		for _, object := range lsRes.Objects {
			if object.Key != path {
				filename := object.Key[len(path):]
				result = append(result, filename)
			}
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			break
		}
	}

	return
}

func (s mockService) bin(platform string) (result []string, err error) {
	return
}

func (s realService) logGetToEnd(reversedID string, fromHead bool, lastAutoID int64) (result []string, newLastAutoID int64, err error) {
	newLastAutoID = lastAutoID

	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = "log"

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("reversed_id", reversedID)
	if fromHead {
		startPK.AddPrimaryKeyColumnWithMinValue("auto_id")
	} else {
		startPK.AddPrimaryKeyColumn("auto_id", lastAutoID+1)
	}

	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("reversed_id", reversedID)
	endPK.AddPrimaryKeyColumnWithMaxValue("auto_id")

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = 10
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := tableStoreClientGlobal.GetRange(getRangeRequest)

	for {
		if err != nil {
			return
		}

		for _, row := range getRangeResp.Rows {
			result = append(result, row.Columns[0].Value.(string))
		}

		if len(getRangeResp.Rows) != 0 {
			newLastAutoID = getRangeResp.Rows[len(getRangeResp.Rows)-1].PrimaryKey.PrimaryKeys[1].Value.(int64)
		}

		if getRangeResp.NextStartPrimaryKey == nil {
			break
		} else {
			getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
			getRangeResp, err = tableStoreClientGlobal.GetRange(getRangeRequest)
		}
	}

	return
}

type task struct {
	topic  string
	status int64
}

//goland:noinspection GoUnusedParameter
func (s mockService) getBriefTaskList(userId string) (result []task, err error) {
	return
}

func (s realService) getBriefTaskList(userId string) (result []task, err error) {
	reversedUserID := reverse(userId)

	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = "task"

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("reversed_user_id", reversedUserID)
	startPK.AddPrimaryKeyColumnWithMaxValue("task_id")

	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("reversed_user_id", reversedUserID)
	endPK.AddPrimaryKeyColumnWithMinValue("task_id")

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.BACKWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = 3
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	getRangeResp, err := tableStoreClientGlobal.GetRange(getRangeRequest)

	if err != nil {
		return
	}

	for _, row := range getRangeResp.Rows {
		result = append(result, task{
			topic:  row.Columns[0].Value.(string),
			status: row.Columns[1].Value.(int64),
		})
	}

	return
}

//goland:noinspection GoUnusedParameter
func (s mockService) logGetToEnd(reversedID string, fromHead bool, lastAutoID int64) (result []string, newLastAutoID int64, err error) {
	return
}

const contentTpl = `{
	"touser":"%s",
	"template_id":"%s",       
	"data":{
		"keyword1":{
			"value":"%s",
			"color":"#173177"
		}
	}
}`

const goodTplID = "F7KKMDk5Cm61PU8XJNAXGEWWFf3UBhEq1F5tsYeMygU"
const badTplID = "3AG4C6YUJBfZ6pt1jwAhzaRd_biqT0vQj9iHmEPgnKc"

type weChatSendMessageRespStruct struct {
	ErrCode uint32 `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s realService) newTask(reversedUserID string, topic string, _type int) (taskID int64, err error) {
	putRowRequest := new(tablestore.PutRowRequest)
	putRowChange := new(tablestore.PutRowChange)
	putRowChange.TableName = "task"

	putPk := new(tablestore.PrimaryKey)
	putPk.AddPrimaryKeyColumn("reversed_user_id", reversedUserID)
	putPk.AddPrimaryKeyColumnWithAutoIncrement("task_id")
	putRowChange.PrimaryKey = putPk

	now := time.Now().Unix()

	putRowChange.AddColumn("topic", topic)
	putRowChange.AddColumn("created_at", now)
	putRowChange.AddColumn("updated_at", now)
	putRowChange.AddColumn("status", int64(0))
	putRowChange.AddColumn("type", int64(_type))

	putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	putRowChange.SetReturnPk()
	putRowRequest.PutRowChange = putRowChange

	putRowResponse, err := tableStoreClientGlobal.PutRow(putRowRequest)
	if err != nil {
		return 0, err
	}

	taskID = putRowResponse.PrimaryKey.PrimaryKeys[1].Value.(int64)

	return
}

func (s mockService) newTask(reversedUserID string, topic string, _type int) (taskID int64, err error) {
	return
}

func (s realService) newLog(taskID int64, content string) (err error) {
	putRowRequest := new(tablestore.PutRowRequest)
	putRowChange := new(tablestore.PutRowChange)
	putRowChange.TableName = "log"

	putPk := new(tablestore.PrimaryKey)

	data := md5.Sum([]byte(strconv.FormatInt(taskID, 10)))
	putPk.AddPrimaryKeyColumn("hash_task_id", fmt.Sprintf("%x", data[:6]))
	putPk.AddPrimaryKeyColumn("task_id", taskID)
	putPk.AddPrimaryKeyColumnWithAutoIncrement("log_id")
	putRowChange.PrimaryKey = putPk

	now := time.Now().Unix()
	putRowChange.AddColumn("content", content)
	putRowChange.AddColumn("created_at", now)

	putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	putRowRequest.PutRowChange = putRowChange

	_, err = tableStoreClientGlobal.PutRow(putRowRequest)

	return
}

func (s mockService) newLog(taskID int64, content string) (err error) {
	return
}

func (s realService) updateTaskStatus(reversedUserID string, taskID int64, status int) (err error) {
	updateRowRequest := new(tablestore.UpdateRowRequest)
	updateRowChange := new(tablestore.UpdateRowChange)
	updateRowChange.TableName = "task"

	updatePk := new(tablestore.PrimaryKey)
	updatePk.AddPrimaryKeyColumn("reversed_user_id", reversedUserID)
	updatePk.AddPrimaryKeyColumn("task_id", taskID)
	updateRowChange.PrimaryKey = updatePk

	now := time.Now().Unix()
	updateRowChange.PutColumn("updated_at", now)
	updateRowChange.PutColumn("status", int64(status))

	updateRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	updateRowRequest.UpdateRowChange = updateRowChange

	_, err = tableStoreClientGlobal.UpdateRow(updateRowRequest)
	if err != nil {
		return
	}

	return
}

func (s mockService) updateTaskStatus(reversedUserID string, taskID int64, status int) (err error) {
	return
}

type weChatGetAccessTokenRespStruct struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint32 `json:"expires_in"`
}

func (s realService) weChatGetAccessToken() (key string, err error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", mpAPPID, mpSECRET)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var wechatResp weChatGetAccessTokenRespStruct
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		log.Println(err)
		return
	}

	if wechatResp.ExpiresIn == 0 {
		err = errors.New("Fail to get Access Token")
		return
	}

	key = wechatResp.AccessToken

	return
}

func (s mockService) weChatGetAccessToken() (key string, err error) {
	return
}
