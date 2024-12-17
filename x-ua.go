package main

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func (s *SSpider) getXUA() error {
	resp, err := s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get("/top/download?os=android")
	if err != nil {
		log.Println("ERROR", err)
		return err
	}

	s.androidXUA = _getXua(resp.String())
	if s.androidXUA == "" {
		log.Println("ERROR", "get x-ua failed")
		return fmt.Errorf("get x-ua failed")
	}

	resp, err = s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get("/top/download?os=ios")
	if err != nil {
		log.Println("ERROR", err)
		return err
	}
	s.iosXUA = _getXua(resp.String())
	if s.iosXUA == "" {
		log.Println("ERROR", "get x-ua failed")
		return fmt.Errorf("get x-ua failed")
	}

	return nil
}

func _getXua(str string) string {
	re := regexp.MustCompile(`const bffRequestUrls = (\[.*?]);`)
	match := re.FindStringSubmatch(str)
	if len(match) < 2 {
		return ""
	}
	// 提取的数组内容
	arrayContent := match[1]                             // 包含 [ 和 ]
	arrayContent = strings.TrimPrefix(arrayContent, "[") // 去掉开头 [
	arrayContent = strings.TrimSuffix(arrayContent, "]") // 去掉结尾 ]
	var urls []string
	for _, item := range strings.Split(arrayContent, ",") {
		_url := strings.TrimSpace(item) // 去掉两端空格
		_url = strings.Trim(_url, `"`)  // 去掉两端引号
		urls = append(urls, _url)
	}

	for _, _url := range urls {
		u, err := url.Parse(_url)
		if err != nil {
			log.Println("ERROR", err)
			return ""
		}
		xua := u.Query().Get("X-UA")
		if xua != "" {
			_xua, err := url.ParseQuery(xua)
			if err != nil {
				log.Println("ERROR", err)
				return ""
			}
			_xua.Del("UID")
			_xua.Del("OSV")
			return url.QueryEscape(_xua.Encode())
		}
	}
	return ""
}
