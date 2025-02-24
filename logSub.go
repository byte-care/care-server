package main

import (
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var subUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func logSub(c *gin.Context) {
	conn, err := subUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		c.String(403, err.Error())
		return
	}
	defer conn.Close()

	id := c.Query("id")
	// todo: check whether task exists

	reversedID := reverse(id)

	getRangeRequest := &tablestore.GetRangeRequest{}
	rangeRowQueryCriteria := &tablestore.RangeRowQueryCriteria{}
	rangeRowQueryCriteria.TableName = "log"

	startPK := new(tablestore.PrimaryKey)
	startPK.AddPrimaryKeyColumn("reversed_task_id", reversedID)
	startPK.AddPrimaryKeyColumnWithMinValue("auto_id")

	endPK := new(tablestore.PrimaryKey)
	endPK.AddPrimaryKeyColumn("reversed_task_id", reversedID)
	endPK.AddPrimaryKeyColumnWithMaxValue("auto_id")

	rangeRowQueryCriteria.StartPrimaryKey = startPK
	rangeRowQueryCriteria.EndPrimaryKey = endPK
	rangeRowQueryCriteria.Direction = tablestore.FORWARD
	rangeRowQueryCriteria.MaxVersion = 1
	rangeRowQueryCriteria.Limit = 50
	getRangeRequest.RangeRowQueryCriteria = rangeRowQueryCriteria

	for {
		getRangeResp, err := tableStoreClientGlobal.GetRange(getRangeRequest)

		for {
			if err != nil {
				c.String(403, err.Error())
				return
			}

			if len(getRangeResp.Rows) == 0 {
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
					return
				}
			} else {
				for _, row := range getRangeResp.Rows {
					err := conn.WriteMessage(websocket.TextMessage, []byte(row.Columns[0].Value.(string)))
					if err != nil {
						c.String(403, err.Error())
						return
					}
				}
			}

			if getRangeResp.NextStartPrimaryKey == nil {
				length := len(getRangeResp.Rows)
				if length == 0 {
					break
				}

				lastIndex := length - 1

				pk := getRangeResp.Rows[lastIndex].PrimaryKey
				pk.PrimaryKeys[1].Value = pk.PrimaryKeys[1].Value.(int64) + 1
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = pk
				break
			} else {
				getRangeRequest.RangeRowQueryCriteria.StartPrimaryKey = getRangeResp.NextStartPrimaryKey
				getRangeResp, err = tableStoreClientGlobal.GetRange(getRangeRequest)
			}
		}

		if err != nil {
			c.String(403, err.Error())
			return
		}

		time.Sleep(2 * time.Second)
	}
}
