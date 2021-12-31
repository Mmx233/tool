package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"reflect"
	"strings"
	"time"
)

type GenTransport struct {
	Timeout   time.Duration
	LocalAddr net.Addr
}

type FullRequest struct {
	Type              string
	Url               string
	Header            map[string]interface{}
	Query             map[string]interface{}
	Body              interface{}
	Cookie            map[string]string
	Redirect          bool
	RedirectCookieJar bool
	Transport         *http.Transport
}

type GetRequest struct {
	Url               string
	Header            map[string]interface{}
	Query             map[string]interface{}
	Cookie            map[string]string
	Redirect          bool
	RedirectCookieJar bool
	Transport         *http.Transport
}

type PostRequest struct {
	Url               string
	Header            map[string]interface{}
	Query             map[string]interface{}
	Body              interface{}
	Cookie            map[string]string
	Redirect          bool
	RedirectCookieJar bool
	Transport         *http.Transport
}

type httP struct { //HTTP操作工具包
	DefaultHeader    map[string]interface{} //默认爬虫header
	DefaultTransport *http.Transport
}

var HTTP = httP{
	DefaultHeader: map[string]interface{}{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36",
	},
	DefaultTransport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Second * 30,
		}).DialContext,
		TLSHandshakeTimeout: time.Second * 30,
	},
}

// 接收指针
func (*httP) fillFullReq(Type string, s interface{}) *FullRequest {
	var r = FullRequest{
		Type: Type,
	}
	v2 := reflect.ValueOf(&r).Elem()
	t := reflect.TypeOf(s).Elem()
	v := reflect.ValueOf(s).Elem()
	for i := 0; i < v.NumField(); i++ {
		v2.FieldByName(t.Field(i).Name).Set(v.Field(i))
	}
	return &r
}

func (a *httP) GenTransport(r *GenTransport) *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   r.Timeout,
			LocalAddr: r.LocalAddr,
		}).DialContext,
		TLSHandshakeTimeout: r.Timeout,
	}
}

// GenRequest 生成请求 底层函数
func (a *httP) GenRequest(Type string, url string, header map[string]interface{}, query map[string]interface{}, body interface{}, cookies map[string]string) (*http.Request, error) {
	//表单
	var form string
	if body != nil {
		if _, ok := header["Content-Type"]; !ok {
			if header == nil {
				header = make(map[string]interface{}, 1)
			}
			header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
		}
		switch {
		case strings.Contains(header["Content-Type"].(string), "x-www-form-urlencoded"):
			var data = make(url2.Values)
			v := reflect.ValueOf(body)
			switch v.Kind() {
			case reflect.Map:
				for _, key := range v.MapKeys() {
					data[fmt.Sprint(key.Interface())] = []string{fmt.Sprint(v.MapIndex(key).Interface())}
				}
			case reflect.Struct:
				t := v.Type()
				for i := 0; i < v.NumField(); i++ {
					data[t.Field(i).Name] = []string{fmt.Sprint(v.Field(i).Interface())}
				}
			default:
				return nil, errors.New("tool http: cannot encode body")
			}
			form = data.Encode()
		case strings.Contains(header["Content-Type"].(string), "json"):
			s, e := json.Marshal(body)
			if e != nil {
				return nil, e
			}
			form = string(s)
		}
	}

	req, err := http.NewRequest(Type, url, strings.NewReader(form))
	if err != nil {
		return nil, err
	}

	//请求头
	if header != nil {
		for k, v := range a.DefaultHeader {
			if _, ok := header[k]; !ok {
				header[k] = v
			}
		}
	} else {
		header = a.DefaultHeader
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

// DefaultReader 执行请求获得io reader的默认流程
func (a *httP) DefaultReader(r *FullRequest) (http.Header, io.ReadCloser, error) {
	req, e := a.GenRequest(r.Type, r.Url, r.Header, r.Query, r.Body, r.Cookie)
	if e != nil {
		return nil, nil, e
	}

	if r.Transport == nil {
		r.Transport = a.DefaultTransport
	}
	var client = &http.Client{
		Transport: r.Transport,
	}

	if !r.Redirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if r.RedirectCookieJar {
		jar, e := cookiejar.New(nil)
		if e != nil {
			return nil, nil, e
		}
		client.Jar = jar
		if r.Cookie != nil {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				u, _ := url2.Parse(r.Url)
				for _, v := range jar.Cookies(u) {
					r.Cookie[v.Name] = v.Value
				}
				return nil
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	return resp.Header, resp.Body, nil
}

// PostReader 执行POST请求，获得io reader
func (a *httP) PostReader(r *PostRequest) (http.Header, io.ReadCloser, error) {
	return a.DefaultReader(a.fillFullReq("POST", r))
}

// GetReader 执行GET请求，获得io reader
func (a *httP) GetReader(r *GetRequest) (http.Header, io.ReadCloser, error) {
	return a.DefaultReader(a.fillFullReq("GET", r))
}

func (*httP) ReadResBodyToByte(i io.ReadCloser) ([]byte, error) {
	defer func() {
		_ = i.Close()
	}()
	return ioutil.ReadAll(i)
}

func (a *httP) ReadResBodyToString(i io.ReadCloser) (string, error) {
	d, e := a.ReadResBodyToByte(i)
	return string(d), e
}

// DecodeResBodyToMap 读取io reader中返回的json写入map
func (a *httP) DecodeResBodyToMap(i io.ReadCloser) (map[string]interface{}, error) {
	var t map[string]interface{}

	//读取
	data, err := a.ReadResBodyToByte(i)
	if err != nil {
		return nil, err
	}

	//解码
	if err = json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return t, nil
}

// Post 表单请求快捷方式
func (a *httP) Post(r *PostRequest) (http.Header, map[string]interface{}, error) {
	d, b, e := a.PostReader(r)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.DecodeResBodyToMap(b)
	return d, c, nil
}

// Get 表单请求快捷方式
func (a *httP) Get(r *GetRequest) (http.Header, map[string]interface{}, error) {
	d, b, e := a.GetReader(r)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.DecodeResBodyToMap(b)
	return d, c, nil
}

func (a *httP) PostBytes(r *PostRequest) (http.Header, []byte, error) {
	d, b, e := a.PostReader(r)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.ReadResBodyToByte(b)
	return d, c, nil
}

func (a *httP) GetBytes(r *GetRequest) (http.Header, []byte, error) {
	d, b, e := a.GetReader(r)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.ReadResBodyToByte(b)
	return d, c, nil
}

func (a *httP) PostString(r *PostRequest) (http.Header, string, error) {
	d, b, e := a.PostReader(r)
	if e != nil {
		return nil, "", e
	}
	c, e := a.ReadResBodyToString(b)
	return d, c, nil
}

func (a *httP) GetString(r *GetRequest) (http.Header, string, error) {
	d, b, e := a.GetReader(r)
	if e != nil {
		return nil, "", e
	}
	c, e := a.ReadResBodyToString(b)
	return d, c, nil
}

func (a httP) DefaultGoquery(r *FullRequest) (*goquery.Document, error) {
	_, resp, e := a.DefaultReader(r)
	if e != nil {
		return nil, e
	}
	d, e := goquery.NewDocumentFromReader(resp)
	_ = resp.Close()
	return d, e
}

func (a httP) GetGoquery(r *GetRequest) (*goquery.Document, error) {
	return a.DefaultGoquery(a.fillFullReq("GET", r))
}

func (a httP) PostGoquery(r *PostRequest) (*goquery.Document, error) {
	return a.DefaultGoquery(a.fillFullReq("POST", r))
}
