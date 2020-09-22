package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Time allowed to write a message to the peer.
const writeWait = 10 * time.Second

// Time allowed to read the next pong message from the peer.
const pongWait = 290 * time.Second

// Send pings to peer with this period. Must be less than pongWait.
const pingPeriod = 260 * time.Second

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func sendPing(quit chan struct{}, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println(err.Error())
			}
		case <-quit:
			return
		}
	}
}

var pubUpgrader = websocket.Upgrader{}

func logPub(c *gin.Context) {
	// Check Signature and Get User(Contain id, secret_key)
	user, err := checkSignature(c, nil)
	if err != nil {
		c.String(403, err.Error())
		return
	}

	conn, err := pubUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.String(403, err.Error())
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	quit := make(chan struct{})
	defer close(quit)

	go sendPing(quit, conn)

	_, topicBytes, err := conn.ReadMessage()
	if err != nil {
		log.Println(err.Error())
		return
	}
	topic := string(topicBytes)

	_, taskTypeBytes, err := conn.ReadMessage()
	if err != nil {
		log.Println(err.Error())
		return
	}

	taskType, err := strconv.Atoi(string(taskTypeBytes))
	if err != nil {
		log.Println(err.Error())
		return
	}

	reversedUserID := reverse(fmt.Sprint(user.ID))

	taskID, err := serviceGlobal.newTask(reversedUserID, topic, taskType)
	if err != nil {
		log.Println(err.Error())
		return
	}

	reversedTaskId := reverse(fmt.Sprint(taskID))

	for {
		_, contentBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				err = serviceGlobal.updateTaskStatus(reversedUserID, taskID, 1)
				if err != nil {
					log.Println(err.Error())
				}

				err = serviceGlobal.email("1423527051@qq.com", topic, "")
				if err != nil {
					log.Println(err.Error())
				}
			} else if websocket.IsCloseError(err, 4000) {
				// Todo: do sth if user want to get notify when exit code not 0
				err = serviceGlobal.updateTaskStatus(reversedUserID, taskID, 2)
				if err != nil {
					log.Println(err.Error())
				}

				err = serviceGlobal.email("1423527051@qq.com", topic, "")
				if err != nil {
					log.Println(err.Error())
				}
			} else {
				// Todo: do sth if user want to get notify when websocket disconnect abnormal
				if err := serviceGlobal.updateTaskStatus(reversedUserID, taskID, 3); err != nil {
					log.Println(err.Error())
				}

				log.Println(err.Error())

				err = serviceGlobal.email("1423527051@qq.com", topic, "")
				if err != nil {
					log.Println(err.Error())
				}
			}

			break
		}

		content := string(contentBytes)

		if err := serviceGlobal.newLog(reversedTaskId, content); err != nil {
			log.Println(err.Error())
		}
	}
}
