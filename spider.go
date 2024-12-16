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
			ImpersonateChrome().
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
	_ = w.Write([]string{"排名", "类型", "ID", "名称", "描述"})

	var num int
	s.getDetail(w, name, args, num)
}

func (s *SSpider) getDetail(w *csv.Writer, name, args string, num int) {
	log.Println("INFO", "req", args)

	url := fmt.Sprintf("%s&X-UA=%s", args, s.androidXUA)
	if strings.Contains(name, "ios") {
		url = fmt.Sprintf("%s&X-UA=%s", url, s.iosXUA)
	}

	resp, err := s.client.R().Get(url)
	if err != nil {
		log.Println("ERROR", err)
		return
	}
	result := gjson.ParseBytes(resp.Bytes())
	if result.Get("success").Bool() != true {
		log.Println("ERROR", "get detail failed")
		return
	}

	list := result.Get("data").Get("list").Array()
	for _, v := range list {
		num += 1
		// 写入文件
		_type := v.Get("type").String()
		_ = w.Write([]string{
			strconv.Itoa(num),
			_type,
			v.Get(_type).Get("id").String(),
			v.Get(_type).Get("title").String(),
			v.Get(_type).Get("rec_text").String(),
		})
	}

	nextPage := result.Get("data").Get("next_page").String()
	if nextPage != "" {
		s.getDetail(w, name, nextPage, num)
	}
}
