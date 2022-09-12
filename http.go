package tool

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
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

type HttpTransportOptions struct {
	Timeout           time.Duration
	LocalAddr         net.Addr
	IdleConnTimeout   time.Duration
	SkipSslCertVerify bool
}

func GenHttpTransport(opt *HttpTransportOptions) *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   opt.Timeout,
			LocalAddr: opt.LocalAddr,
		}).DialContext,
		TLSHandshakeTimeout: opt.Timeout,
		IdleConnTimeout:     opt.IdleConnTimeout,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: opt.SkipSslCertVerify},
	}
}

type HttpClientOptions struct {
	Transport *http.Transport
	//禁止跟随重定向
	NoRedirect bool
	//启用 cookiejar
	RedirectCookieJar bool
	//超时时间
	Timeout time.Duration
}

func GenHttpClient(opt *HttpClientOptions) *http.Client {
	c := &http.Client{
		Transport: opt.Transport,
		Timeout:   opt.Timeout,
	}

	if opt.NoRedirect {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if opt.RedirectCookieJar {
		jar, _ := cookiejar.New(nil)
		c.Jar = jar
	}

	return c
}

func NewHttpTool(c *http.Client) *Http {
	if c == nil {
		c = http.DefaultClient
	}
	return &Http{
		Client: c,
	}
}

type Http struct {
	Client *http.Client
}

type DoHttpReq struct {
	Url    string
	Header map[string]interface{}
	Query  map[string]interface{}
	Body   interface{}
	Cookie map[string]string
}

func (a *Http) GenReq(Type string, opt *DoHttpReq) (*http.Request, error) {
	//表单
	var form io.Reader
	if opt.Body != nil {
		if i, ok := opt.Body.(io.Reader); ok {
			form = i
		} else {
			var body bytes.Buffer
			form = &body
			v := reflect.ValueOf(opt.Body)
			if _, ok := opt.Header["Content-Type"]; !ok {
				if opt.Header == nil {
					opt.Header = make(map[string]interface{}, 1)
				}
				switch v.Kind() {
				case reflect.Ptr:
				case reflect.Struct:
					opt.Header["Content-Type"] = "application/json; charset=utf-8"
				case reflect.Map:
					opt.Header["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
				default:
					return nil, errors.New("tool http: unknown body type")
				}
			}
			switch {
			case strings.Contains(opt.Header["Content-Type"].(string), "x-www-form-urlencoded"):
				var data = make(url2.Values)
				switch v.Kind() {
				case reflect.Map:
					for _, key := range v.MapKeys() {
						data[fmt.Sprint(key.Interface())] = []string{fmt.Sprint(v.MapIndex(key).Interface())}
					}
				default:
					return nil, errors.New("tool http: cannot encode body")
				}
				body.Write([]byte(data.Encode()))
			case strings.Contains(opt.Header["Content-Type"].(string), "json"):
				s, e := json.Marshal(opt.Body)
				if e != nil {
					return nil, e
				}
				body.Write(s)
			}
		}
	}

	req, err := http.NewRequest(Type, opt.Url, form)
	if err != nil {
		return nil, err
	}

	for k, v := range opt.Header {
		req.Header.Add(k, fmt.Sprint(v))
	}

	//url参数
	q := req.URL.Query()
	for k, v := range opt.Query {
		q.Add(k, fmt.Sprint(v))
	}
	req.URL.RawQuery = q.Encode()

	//cookie
	for k, v := range opt.Cookie {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	return req, nil
}

func (a *Http) Request(Type string, opt *DoHttpReq) (*http.Response, error) {
	req, e := a.GenReq(Type, opt)
	if e != nil {
		return nil, e
	}

	return a.Client.Do(req)
}

func (a *Http) PostRequest(opt *DoHttpReq) (*http.Response, error) {
	return a.Request("POST", opt)
}

func (a *Http) GetRequest(opt *DoHttpReq) (*http.Response, error) {
	return a.Request("GET", opt)
}

func (*Http) ReadResBodyToByte(i io.ReadCloser) ([]byte, error) {
	defer i.Close()
	return ioutil.ReadAll(i)
}

func (a *Http) ReadResBodyToString(i io.ReadCloser) (string, error) {
	d, e := a.ReadResBodyToByte(i)
	return string(d), e
}

func (a *Http) UnMarshalResBodyToMap(i io.ReadCloser) (map[string]interface{}, error) {
	defer i.Close()
	var t map[string]interface{}
	return t, json.NewDecoder(i).Decode(&t)
}

// Post 表单请求快捷方式
func (a *Http) Post(opt *DoHttpReq) (*http.Response, map[string]interface{}, error) {
	res, e := a.PostRequest(opt)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.UnMarshalResBodyToMap(res.Body)
	return res, c, e
}

// Get 表单请求快捷方式
func (a *Http) Get(opt *DoHttpReq) (*http.Response, map[string]interface{}, error) {
	res, e := a.GetRequest(opt)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.UnMarshalResBodyToMap(res.Body)
	return res, c, e
}

func (a *Http) PostBytes(opt *DoHttpReq) (*http.Response, []byte, error) {
	res, e := a.PostRequest(opt)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.ReadResBodyToByte(res.Body)
	return res, c, e
}

func (a *Http) GetBytes(opt *DoHttpReq) (*http.Response, []byte, error) {
	res, e := a.GetRequest(opt)
	if e != nil {
		return nil, nil, e
	}
	c, e := a.ReadResBodyToByte(res.Body)
	return res, c, e
}

func (a *Http) PostString(opt *DoHttpReq) (*http.Response, string, error) {
	res, e := a.PostRequest(opt)
	if e != nil {
		return nil, "", e
	}
	c, e := a.ReadResBodyToString(res.Body)
	return res, c, e
}

func (a *Http) GetString(opt *DoHttpReq) (*http.Response, string, error) {
	res, e := a.GetRequest(opt)
	if e != nil {
		return nil, "", e
	}
	c, e := a.ReadResBodyToString(res.Body)
	return res, c, e
}

type HttpRequest struct {
	c   *Http
	Req *http.Request
}

func (a *Http) PrepareRequest(Type string, opt *DoHttpReq) (*HttpRequest, error) {
	req, e := a.GenReq(Type, opt)
	if e != nil {
		return nil, e
	}

	return &HttpRequest{
		c:   a,
		Req: req,
	}, nil
}

func (a *HttpRequest) Do() (*http.Response, error) {
	return a.c.Client.Do(a.Req)
}
