package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tidwall/gjson"
)

func (s *SSpider) getTerms() error {
	resp, err := s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get(fmt.Sprintf("/webapiv2/top/v3/terms?X-UA=%s", s.androidXUA))
	if err != nil {
		log.Println("ERROR", err)
		return err
	}
	result := gjson.ParseBytes(resp.Bytes())
	if result.Get("success").Bool() != true {
		log.Println("ERROR", "get terms failed")
		return fmt.Errorf("get terms failed")
	}
	list := result.Get("data").Get("list").Array()
	for _, v := range list {
		s.terms["android_"+v.Get("label").String()] = v.Get("url").String()
	}

	resp, err = s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get(fmt.Sprintf("/webapiv2/top/v3/terms?X-UA=%s", s.iosXUA))
	if err != nil {
		log.Println("ERROR", err)
		return err
	}

	result = gjson.ParseBytes(resp.Bytes())
	if result.Get("success").Bool() != true {
		log.Println("ERROR", "get terms failed")
		return fmt.Errorf("get terms failed")
	}
	list = result.Get("data").Get("list").Array()
	for _, v := range list {
		s.terms["ios_"+v.Get("label").String()] = v.Get("url").String()
	}
	return nil
}
