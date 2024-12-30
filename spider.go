package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

type SSpider struct {
	wg     *sync.WaitGroup
	client *req.Client

	androidXUA string
	iosXUA     string

	terms map[string]string
}

func NewSpider() *SSpider {
	s := &SSpider{
		wg: new(sync.WaitGroup),
		client: req.NewClient().
			EnableHTTP3().
			//EnableDebugLog().
			ImpersonateChrome().
			SetTLSFingerprintChrome().
			SetBaseURL("https://www.taptap.cn/").
			SetCommonHeaders(map[string]string{
				"Accept-Encoding": "identity",
				"Referer":         "https://www.taptap.cn/",
			}),
		terms: make(map[string]string),
	}
	return s
}

func (s *SSpider) Run() (err error) {
	if err = s.getXUA(); err != nil {
		return err
	}

	if err = s.getTerms(); err != nil {
		return err
	}

	for name, args := range s.terms {
		s.wg.Add(1)
		go s.worker(name, args)
	}

	s.wg.Wait()

	log.Println("INFO", "spider done")
	return nil
}

func (s *SSpider) worker(name, args string) {
	defer s.wg.Done()

	file, err := os.Create(filepath.Join("output", name+".csv"))
	if err != nil {
		s.wg.Done()
		log.Println("ERROR", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, _ = file.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(file)
	defer w.Flush()
	_ = w.Write([]string{"排名", "评分", "下载", "关注", "类型", "ID", "名称", "描述", "信息", "标签"})

	var num int
	s.getList(w, name, args, num)
}

func (s *SSpider) getList(w *csv.Writer, name, args string, num int) {
	log.Println("INFO", "req", args)

	ua := fmt.Sprintf("X-UA=%s", s.androidXUA)
	if strings.Contains(name, "ios") {
		ua = fmt.Sprintf("X-UA=%s", s.iosXUA)
	}
	url := fmt.Sprintf("%s&%s", args, ua)
	resp, err := s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get(url)
	if err != nil {
		log.Println("ERROR", err)
		return
	}
	result := gjson.ParseBytes(resp.Bytes())
	if result.Get("success").Bool() != true {
		log.Println("ERROR", "get detail failed", result.String())
		return
	}

	list := result.Get("data").Get("list").Array()
	for _, v := range list {
		num += 1
		_type := v.Get("type").String()
		id := v.Get(_type).Get("id").Int()

		detailUrl := fmt.Sprintf("os=android&X-UA=%s", s.androidXUA)
		if strings.Contains(name, "ios") {
			detailUrl = fmt.Sprintf("os=ios&X-UA=%s", s.iosXUA)
		}
		switch _type {
		case "app":
			detailUrl = fmt.Sprintf("/webapiv2/app/v4/detail?id=%d&%s", id, detailUrl)
		case "craft":
			detailUrl = fmt.Sprintf("/webapiv2/craft/v1/detail-by-id?id=%d&%s", id, detailUrl)
		default:
			log.Println("ERROR", "unknown type", _type, id, url)
		}
		s.getDetail(w, detailUrl, _type, num)
	}

	nextPage := result.Get("data").Get("next_page").String()
	if nextPage != "" {
		s.getList(w, name, nextPage, num)
	}
}

func (s *SSpider) getDetail(w *csv.Writer, url, _type string, num int) {
	resp, err := s.client.R().
		SetRetryCount(-1).
		SetRetryBackoffInterval(1*time.Second, 5*time.Second).
		Get(url)
	if err != nil {
		log.Println("ERROR", err)
		return
	}
	result := gjson.ParseBytes(resp.Bytes())
	if result.Get("success").Bool() != true {
		log.Println("ERROR", "get detail failed", url, result.String())
		return
	}
	//_ = w.Write([]string{"排名", "评分", "下载", "关注", "类型", "ID", "名称", "描述", "信息", "标签"})
	information := func() string {
		array := result.Get("data").Get("information").Array()
		if len(array) == 0 {
			array = result.Get("data").Get("compliance_info").Get("list").Array()
		}
		var res []string
		for _, v := range array {
			res = append(res, v.Get("title").String()+": "+v.Get("text").String())
		}
		return strings.Join(res, "\n")
	}()
	tags := func() string {
		array := result.Get("data").Get("tags").Array()
		var res []string
		for _, v := range array {
			res = append(res, v.Get("value").String())
		}
		return strings.Join(res, ", ")
	}()
	hitsTotal := func() string {
		count := result.Get("data").Get("stat").Get("hits_total").String()
		if count == "" {
			count = result.Get("data").Get("stat").Get("played_count").String()
		}
		return count
	}()
	_ = w.Write([]string{
		strconv.Itoa(num),
		result.Get("data").Get("stat").Get("rating").Get("score").String(),
		hitsTotal,
		result.Get("data").Get("stat").Get("fans_count").String(),
		_type,
		result.Get("data").Get("id").String(),
		result.Get("data").Get("title").String(),
		result.Get("data").Get("description").Get("text").String(),
		information,
		tags,
	})
}
