package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"

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

	err = boltDB.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte("device"))

		c := b.Cursor()
		for deviceToken, _ := c.First(); deviceToken != nil; deviceToken, _ = c.Next() {
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

func akRegister(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	err := r.ParseForm()
	if err != nil {
		logrus.Error(err)
	}

	var deviceToken string
	for key, value := range r.Form {
		if strings.ToLower(key) == "devicetoken" {
			deviceToken = value[0]
			break
		}
	}

	if len(deviceToken) <= 0 {
		_, err = fmt.Fprint(w, responseString(400, "deviceToken 不能为空"))
		if err != nil {
			logrus.Error(err)
		}
		return
	}

	err = boltDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("device"))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(deviceToken), []byte(time.Now().Local().String()))
	})

	if err != nil {
		_, err = fmt.Fprint(w, responseString(400, "注册设备失败"))
		if err != nil {
			logrus.Error(err)
		}
		return
	}
	logrus.Info("注册设备成功")
	logrus.Info("deviceToken: ", deviceToken)
	_, err = fmt.Fprint(w, responseData(200, map[string]interface{}{"key": "do-not-use"}, "注册成功"))
	if err != nil {
		logrus.Error(err)
	}
}
