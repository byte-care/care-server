package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/byte-care/care-server-core/model"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"

	core "github.com/byte-care/care-core"
)

type notifyService interface {
	logPubNormal(userID uint, topic string) (err error)
	logPubExitAbnormal(userID uint, topic string) (err error)
	logPubDisconnectAbnormal(userID uint, topic string) (err error)
	getChannelID(userID uint) (address string, err error)
}

type realEmailNotifyService struct {
}

func (r realEmailNotifyService) getChannelID(userID uint) (address string, err error) {
	var cEmail model.ChannelEmail
	result := db.Select("address").Where("user_id = ?", userID).First(&cEmail)
	err = result.Error
	if err != nil {
		return
	}

	address = cEmail.Address
	return
}

func (r realEmailNotifyService) logPubNormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	err = serviceGlobal.email(cid, topic, "logPubNormal")
	return
}

func (r realEmailNotifyService) logPubExitAbnormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	err = serviceGlobal.email(cid, topic, "logPubExitAbnormal")
	return
}

func (r realEmailNotifyService) logPubDisconnectAbnormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	err = serviceGlobal.email(cid, topic, "logPubDisconnectAbnormal")
	return
}

type mockEmailNotifyService struct {
}

func (m mockEmailNotifyService) getChannelID(userID uint) (address string, err error) {
	panic("implement me")
}

func (m mockEmailNotifyService) logPubNormal(userID uint, topic string) (err error) {
	panic("implement me")
}

func (m mockEmailNotifyService) logPubExitAbnormal(userID uint, topic string) (err error) {
	panic("implement me")
}

func (m mockEmailNotifyService) logPubDisconnectAbnormal(userID uint, topic string) (err error) {
	panic("implement me")
}

type realWechatNotifyService struct {
}

func (r realWechatNotifyService) getChannelID(userID uint) (openID string, err error) {
	var cWeChat model.ChannelWeChat
	result := db.Select("mp_open_id").Where("user_id = ?", userID).First(&cWeChat)
	err = result.Error
	if err != nil {
		return
	}

	openID = cWeChat.MPOpenID
	return
}

func (r realWechatNotifyService) logPubNormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	content := fmt.Sprintf(contentTpl, cid, "✔ 执行成功", topic, "完成", "刚刚", "")
	err = serviceGlobal.weChat(content)
	return
}

func (r realWechatNotifyService) logPubExitAbnormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	content := fmt.Sprintf(contentTpl, cid, "❌ 执行失败", topic, "程序异常退出", "刚刚", "")
	err = serviceGlobal.weChat(content)
	return
}

func (r realWechatNotifyService) logPubDisconnectAbnormal(userID uint, topic string) (err error) {
	cid, err := r.getChannelID(userID)
	if err != nil {
		return
	}

	content := fmt.Sprintf(contentTpl, cid, "❌ 执行失败", topic, "连接异常断开", "刚刚", "")
	err = serviceGlobal.weChat(content)
	return
}

type mockWechatNotifyService struct {
}

func (m mockWechatNotifyService) getChannelID(userID uint) (address string, err error) {
	panic("implement me")
}

func (m mockWechatNotifyService) logPubNormal(userID uint, topic string) (err error) {
	panic("implement me")
}

func (m mockWechatNotifyService) logPubExitAbnormal(userID uint, topic string) (err error) {
	panic("implement me")
}

func (m mockWechatNotifyService) logPubDisconnectAbnormal(userID uint, topic string) (err error) {
	panic("implement me")
}

type service interface {
	email(address string, subject string, body string) error
	weChat(content string) error
	sms(number string, code string) error
	bin(platform string) (string, error)
	getBriefTaskList(userId string) (result []task, err error)
	newTask(reversedUserID string, topic string, _type int) (taskID int64, err error)
	updateTaskStatus(reversedUserID string, taskID int64, status int) (err error)
	newLog(reversedTaskId string, content string) (err error)
	weChatGetAccessToken() (string, error)
}

type realService struct {
}

func (s realService) weChat(content string) (err error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", mpAccessToken)

	req, err := http.NewRequest("POST", url, strings.NewReader(content))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
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

	var wechatResp weChatSendMessageRespStruct
	err = json.Unmarshal(body, &wechatResp)
	if err != nil {
		log.Println(err)
		return
	}

	if wechatResp.ErrCode != 0 {
		err = errors.New("Fail to send Message")
		log.Println(wechatResp.ErrMsg)
		return
	}

	return
}

type mockService struct {
}

func (s mockService) weChat(content string) error {
	panic("implement me")
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

func (s realService) bin(platform string) (result string, err error) {
	bucketName := "care-bin"

	bucket, err := ossClientGlobal.Bucket(bucketName)
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/", platform)
	pathLen := len(path)

	marker := ""

	lsRes, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(path))
	if err != nil {
		return
	}

	if len(lsRes.Objects) != 1 {
		return "", errors.New("more than one version")
	}

	result = lsRes.Objects[0].Key[pathLen:]
	return
}

func (s mockService) bin(platform string) (result string, err error) {
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
			topic:  row.Columns[2].Value.(string),
			status: row.Columns[1].Value.(int64),
		})
	}

	return
}

const contentTpl = `{
	"touser":"%s",
	"template_id":"yklsI1iWhuOsGsXdraYXROMYLfXoiFMXH6FGaR1kqUE",       
	"data":{
		"first":{
			"value":"%s",
			"color":"#173177"
		},
		"keyword1":{
			"value":"%s",
			"color":"#173177"
		},
		"keyword2":{
			"value":"%s",
			"color":"#173177"
		},
		"keyword3":{
			"value":"%s",
			"color":"#173177"
		},
		"remark":{
			"value":"%s",
			"color":"#173177"
		}
	}
}`

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
	putRowChange.AddColumn("status", int64(0))
	putRowChange.AddColumn("type", int64(_type))
	putRowChange.AddColumn("created_at", now)
	putRowChange.AddColumn("updated_at", now)

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

func (s realService) newLog(reversedTaskId string, content string) (err error) {
	putRowRequest := new(tablestore.PutRowRequest)
	putRowChange := new(tablestore.PutRowChange)
	putRowChange.TableName = "log"

	putPk := new(tablestore.PrimaryKey)

	putPk.AddPrimaryKeyColumn("reversed_task_id", reversedTaskId)
	putPk.AddPrimaryKeyColumnWithAutoIncrement("auto_id")
	putRowChange.PrimaryKey = putPk

	now := time.Now().Unix()
	putRowChange.AddColumn("content", content)
	putRowChange.AddColumn("created_at", now)

	putRowChange.SetCondition(tablestore.RowExistenceExpectation_IGNORE)
	putRowRequest.PutRowChange = putRowChange

	_, err = tableStoreClientGlobal.PutRow(putRowRequest)

	return
}

func (s mockService) newLog(reversedTaskId string, content string) (err error) {
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
