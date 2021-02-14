package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
)

func akFeed(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logrus.Error(err)
	}

	params := make(map[string]interface{})
	for key, value := range r.Form {
		params[strings.ToLower(key)] = value[0]
	}

	logrus.Println(" ========================== ")
	logrus.Println("params: ", params)
	logrus.Println(" ========================== ")

	err = boltDB.View(func(t *bbolt.Tx) error {
		b := t.Bucket([]byte("device"))

		c := b.Cursor()
		for k, deviceToken := c.First(); k != nil; k, deviceToken = c.Next() {
			go akPush(akPayload(params), string(deviceToken))
		}

		return nil
	})

	_, err = fmt.Fprint(w, responseString(200, ""))
	if err != nil {
		logrus.Error(err)
	}
}

func akPayload(params map[string]interface{}) *payload.Payload {
	var sound string = "1107"
	pl := payload.NewPayload().Sound(sound).Category("myNotificationCategory")

	pl = pl.Custom("url", params["url"])
	pl.AlertTitle(params["title"].(string))
	pl.AlertBody(params["body"].(string))

	return pl.MutableContent()
}

func akPush(pl *payload.Payload, deviceToken string) error {
	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken

	notification.Payload = pl
	notification.Topic = "me.fin.bark"
	res, err := apnsClient.Push(notification)

	if err != nil {
		logrus.Errorf("Error:", err)
		return fmt.Errorf("与苹果推送服务器传输数据失败: %w", err)
	}
	logrus.Infof("%v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
	if res.StatusCode == 200 {
		return nil
	} else {
		return errors.New("推送发送失败 " + res.Reason)
	}
}
