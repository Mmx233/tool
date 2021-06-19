package tool

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"strings"
)

type httP struct { //HTTP操作工具包
	defaultHeader map[string]string //默认爬虫header
}

var HTTP = httP{
	defaultHeader: map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7",
		"Cache-Control":   "max-age=0",
		"Connection":      "keep-alive",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
	},
}

// GenRequest 生成请求 底层函数
func (a *httP) GenRequest(Type string, url string, header map[string]interface{}, query map[string]interface{}, body map[string]interface{}, cookies map[string]string) (*http.Request, error) {
	//表单
	var form string
	if body != nil {
		var data = make(url2.Values)
		for k, v := range body {
			data[k] = []string{fmt.Sprint(v)}
		}
		form = data.Encode()
		if header == nil {
			header = make(map[string]interface{}, 1)
		}
		header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
	}

	req, err := http.NewRequest(Type, url, strings.NewReader(form))
	if err != nil {
		return nil, err
	}

	//请求头
	for k, v := range a.defaultHeader {
		req.Header.Add(k, v)
	}
	for k, v := range header {
		req.Header.Add(k, fmt.Sprint(v))
	}

	//url参数
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, fmt.Sprint(v))
	}
	req.URL.RawQuery = q.Encode()

	//cookie
	for k, v := range cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	return req, nil
}

//执行请求获得io reader的默认流程
func (a *httP) defaultReader(Type string, url string, header map[string]interface{}, query map[string]interface{}, body map[string]interface{}, cookies map[string]string) (http.Header, io.ReadCloser, error) {
	req, err := a.GenRequest(Type, url, header, query, body, cookies)
	if err != nil {
		return nil, nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	return resp.Header, resp.Body, nil
}

// GetReader 执行GET请求，获得io reader
func (a *httP) GetReader(url string, header map[string]interface{}, query map[string]interface{}, cookie map[string]string) (http.Header, io.ReadCloser, error) {
	return a.defaultReader("GET", url, header, query, nil, cookie)
}

// PostReader 执行POST请求，获得io reader
func (a *httP) PostReader(url string, header map[string]interface{}, query map[string]interface{}, body map[string]interface{}, cookie map[string]string) (http.Header, io.ReadCloser, error) {
	return a.defaultReader("POST", url, header, query, body, cookie)
}

// DecodeResBodyToMap 读取io reader中返回的json写入map
func (a *httP) DecodeResBodyToMap(i io.ReadCloser) (map[string]interface{}, error) {
	var t map[string]interface{}

	//读取
	data, err := ioutil.ReadAll(i)
	if err != nil {
		return nil, err
	}

	//解码
	if err = json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return t, nil
}

// POST 表单请求快捷方式
func (a *httP) POST(url string, header map[string]interface{}, query map[string]interface{}, body map[string]interface{}, cookie map[string]string) (http.Header, map[string]interface{}, error) {
	d, b, e := a.PostReader(url, header, query, body, cookie)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.DecodeResBodyToMap(b)
	_ = b.Close()
	if e != nil {
		return nil, nil, e
	}
	return d, c, nil
}

// Get 表单请求快捷方式
func (a *httP) Get(url string, header map[string]interface{}, query map[string]interface{}, cookie map[string]string) (http.Header, map[string]interface{}, error) {
	d, b, e := a.GetReader(url, header, query, cookie)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.DecodeResBodyToMap(b)
	_ = b.Close()
	if e != nil {
		return nil, nil, e
	}
	return d, c, nil
}

// GetLocation 获取301/302目标地址
func (a httP) GetLocation(url string, header map[string]interface{}, query map[string]interface{}, cookie map[string]string) (string, error) {
	req, e := a.GenRequest("GET", url, header, query, nil, cookie)
	if e != nil {
		return "", e
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, e := client.Do(req)
	if e != nil {
		return "", e
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return resp.Header.Get("Location"), nil
}

// GetGoquery 获取goquery
func (a httP) GetGoquery(url string, header map[string]interface{}, query map[string]interface{}, cookie map[string]string) (*goquery.Document, error) {
	_, resp, e := a.defaultReader("GET", url, header, query, nil, cookie)
	if e != nil {
		return nil, e
	}
	defer func(resp io.ReadCloser) {
		_ = resp.Close()
	}(resp)
	d, e := goquery.NewDocumentFromReader(resp)
	if e != nil {
		return nil, e
	}
	return d, nil
}
