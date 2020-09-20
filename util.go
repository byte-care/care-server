package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	core "github.com/byte-care/care-core"
	"github.com/byte-care/care-server-core/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var aliyunRegionID string
var aliyunAccessKey string
var aliyunSecretKey string

var mpAPPID string
var mpSECRET string
var mpAccessToken string

var GitHubClientID string
var GitHubClientSecret string

var clientGlobal *sdk.Client
var ossClientGlobal *oss.Client
var tableStoreClientGlobal *tablestore.TableStoreClient

var privateKeyGlobal *rsa.PrivateKey
var secretKeyStr string
var secretKeyGlobal []byte

var db *gorm.DB
var serviceGlobal service
var wechatNotifyServiceGlobal notifyService

func autoMigrate() {
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.ChannelEmail{})
	db.AutoMigrate(&model.ChannelWeChat{})
	db.AutoMigrate(&model.OAuthGitHub{})
}

func setMPAccessToken() {
	accessToken, err := serviceGlobal.weChatGetAccessToken()
	if err != nil {
		log.Println(err)
		return
	}

	mpAccessToken = accessToken
}

func setup(test bool) {
	if !test {
		aliyunRegionIDLocal, ok := os.LookupEnv("CARE_ALIYUN_REGION_ID")
		if !ok {
			panic("CARE_ALIYUN_REGION_ID not set")
		}
		aliyunRegionID = aliyunRegionIDLocal

		aliyunAccessKeyLocal, ok := os.LookupEnv("CARE_ALIYUN_ACCESS_KEY")
		if !ok {
			panic("CARE_ALIYUN_ACCESS_KEY not set")
		}
		aliyunAccessKey = aliyunAccessKeyLocal

		aliyunSecretKeyLocal, ok := os.LookupEnv("CARE_ALIYUN_SECRET_KEY")
		if !ok {
			panic("CARE_ALIYUN_SECRET_KEY not set")
		}
		aliyunSecretKey = aliyunSecretKeyLocal

		secretKeyStrLocal, ok := os.LookupEnv("CARE_SECRET_KEY_STR")
		if !ok {
			panic("CARE_SECRET_KEY_STR not set")
		}
		secretKeyStr = secretKeyStrLocal

		GitHubClientIDLocal, ok := os.LookupEnv("CARE_GITHUB_CLIENT_ID")
		if !ok {
			panic("CARE_GITHUB_CLIENT_ID not set")
		}
		GitHubClientID = GitHubClientIDLocal

		GitHubClientSecretLocal, ok := os.LookupEnv("CARE_GITHUB_CLIENT_SECRET")
		if !ok {
			panic("CARE_GITHUB_CLIENT_SECRET not set")
		}
		GitHubClientSecret = GitHubClientSecretLocal

		mpAPPIDLocal, ok := os.LookupEnv("CARE_MP_APPID")
		if !ok {
			panic("CARE_MP_APPID not set")
		}
		mpAPPID = mpAPPIDLocal

		mpSECRETLocal, ok := os.LookupEnv("CARE_MP_SECRET")
		if !ok {
			panic("CARE_MP_SECRET not set")
		}
		mpSECRET = mpSECRETLocal
	}

	mrand.Seed(time.Now().UnixNano())

	// Set Private Key
	privateKeyLocal, err := getPrivateKey(test)
	if err != nil {
		panic(err)
	}

	privateKeyGlobal = privateKeyLocal

	secretKeyLocal, err := base64.URLEncoding.DecodeString(secretKeyStr)
	if err != nil {
		panic(err)
	}

	secretKeyGlobal = secretKeyLocal

	// Connect DB
	dbLocal, err := connectDB(test)
	if err != nil {
		panic(err)
	}

	db = dbLocal
	autoMigrate()

	// Init Aliyun Client
	clientLocal, err := sdk.NewClientWithAccessKey(aliyunRegionID, aliyunAccessKey, aliyunSecretKey)
	if err != nil {
		panic(err)
	}

	clientGlobal = clientLocal

	// Init OSS Client
	ossClientLocal, err := oss.New("oss-cn-beijing-internal.aliyuncs.com", aliyunAccessKey, aliyunSecretKey)
	if err != nil {
		panic(err)
	}

	ossClientGlobal = ossClientLocal

	// Init TableStore Client
	tableStoreClientLocal := tablestore.NewClient("https://kan.cn-beijing.vpc.tablestore.aliyuncs.com", "kan", aliyunAccessKey, aliyunSecretKey)
	tableStoreClientGlobal = tableStoreClientLocal
}

func connectDB(test bool) (db *gorm.DB, err error) {
	if test {
		db, err = gorm.Open("sqlite3", "DB.db")
	} else {
		user, ok := os.LookupEnv("WP_RDS_ACCOUNT_NAME")
		if !ok {
			return nil, errors.New("WP_RDS_ACCOUNT_NAME not set")
		}

		password, ok := os.LookupEnv("WP_RDS_ACCOUNT_PASSWORD")
		if !ok {
			return nil, errors.New("WP_RDS_ACCOUNT_PASSWORD not set")
		}

		address, ok := os.LookupEnv("WP_RDS_CONNECTION_ADDRESS")
		if !ok {
			return nil, errors.New("WP_RDS_CONNECTION_ADDRESS not set")
		}

		dbname, ok := os.LookupEnv("WP_RDS_DATABASE")
		if !ok {
			return nil, errors.New("WP_RDS_DATABASE not set")
		}

		dsn := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, address, dbname)
		db, err = gorm.Open("mysql", dsn)
	}

	return
}

type codeClaims struct {
	CodeHash  string `json:"code_hash"`
	ChannelID string `json:"channel_id"`
	jwt.StandardClaims
}

type idClaims struct {
	ID string `json:"id"`
	jwt.StandardClaims
}

func generateKey() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", errors.New("Can't generate key")
	}

	key := base64.URLEncoding.EncodeToString(randomBytes)

	return key, nil
}

func getPrivateKey(test bool) (*rsa.PrivateKey, error) {
	if test {
		reader := rand.Reader
		bitSize := 512

		return rsa.GenerateKey(reader, bitSize)
	}

	url, ok := os.LookupEnv("CARE_PRIVATE_KEY_URL")
	if !ok {
		return nil, errors.New("CARE_PRIVATE_KEY_URL not set")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil

}

func generateCode(channelID string) (raw string, token string, err error) {
	ints := make([]string, 6)

	for i := 0; i <= 5; i++ {
		v := mrand.Intn(10)
		ints[i] = strconv.Itoa(v)
	}

	raw = strings.Join(ints, "")
	hash := core.HashString(raw, secretKeyGlobal)

	token, err = generateCodeToken(hash, channelID)
	if err != nil {
		return "", "", err
	}

	return
}

func generateIDToken(id string) (tokenString string, err error) {
	claims := idClaims{
		id,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 1, 0).Unix(),
			Issuer:    "bytecare.xyz",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err = token.SignedString(privateKeyGlobal)
	if err != nil {
		return "", err
	}

	return
}

func generateCodeToken(codeHash string, channelID string) (tokenString string, err error) {
	claims := codeClaims{
		codeHash,
		channelID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			Issuer:    "bytecare.xyz",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err = token.SignedString(privateKeyGlobal)
	if err != nil {
		return "", err
	}

	return
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func checkSignature(c *gin.Context, specificParameter map[string]string) (*model.User, error) {

	signatureNonce := c.GetHeader("Care-Nonce")
	if signatureNonce == "" {
		return nil, errors.New("No SignatureNonce")
	}

	timestamp := c.GetHeader("Care-Timestamp")
	if timestamp == "" {
		return nil, errors.New("No Timestamp")
	}

	accessKey := c.GetHeader("Care-Key")
	if accessKey == "" {
		return nil, errors.New("No AccessKey")
	}

	signature := c.GetHeader("Care-Signature")
	if signature == "" {
		return nil, errors.New("No Signature")
	}

	commonParameter := core.CommonParameter{
		AccessKey:      accessKey,
		SignatureNonce: signatureNonce,
		Timestamp:      timestamp,
	}

	var user model.User
	db.Select("id, secret_key").Where("access_key = ?", accessKey).First(&user)
	if user.ID == 0 {
		return nil, errors.New("User not Exist")
	}

	credential, err := core.NewCredential(accessKey, user.SecretKey)
	if err != nil {
		return nil, err
	}

	s := credential.Sign(commonParameter, specificParameter)
	if s != signature {
		return nil, errors.New("Signature not Valid")
	}

	return &user, nil
}
