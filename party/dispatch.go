package party

import (
	"crypto/tls"
	"github.com/shopspring/decimal"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	fc *fasthttp.Client
)

const (
	apiTimeOut = time.Second * 12
)

func ParamEncode(args map[string]string) string {
	if len(args) < 1 {
		return ""
	}

	data := url.Values{}
	for k, v := range args {
		data.Set(k, v)
	}
	return data.Encode()
}

func PushLog(requestBody []byte, username, name, requestURI string, code int, body []byte, err error) {

	//tag := "platform"
	//l := PlatLog{
	//	Requesturl:  requestURI,
	//	Requestbody: string(requestBody),
	//	Statuscode:  code,
	//	Body:        string(body),
	//	Err:         "",
	//	Level:       "info",
	//	Name:        name,
	//	Username:    username,
	//}
	//if err != nil {
	//	l.Err = err.Error()
	//	l.Level = "warm"
	//}
	//if code != fasthttp.StatusCreated && code != fasthttp.StatusOK {
	//	l.Level = "warm"
	//}
	//
	//if l.Level == "warm" {
	//	//debug.PrintStack()
	//	fmt.Println("PushLog Requesturl = ", requestURI)
	//}

	//err = Meta.Zlog.Post(Meta.EsPrefix+tag, l)
	//if err != nil {
	//	fmt.Printf("Push Platform Log is error: %s \n", err.Error())
	//}
}

func HttpPostWithPushLog(requestBody []byte, username, name, requestURI string, headers map[string]string) (int, []byte, error) {
	statusCode, body, err := httpPost(requestBody, requestURI, headers)
	PushLog(requestBody, username, name, requestURI, statusCode, body, err)
	return statusCode, body, err
}

func HttpGetWithPushLog(username, name, requestURI string) (int, []byte, error) {
	statusCode, body, err := httpGet(requestURI)
	PushLog(nil, username, name, requestURI, statusCode, body, err)
	return statusCode, body, err
}

func HttpGetHeaderWithLog(username, name, requestURI string, headers map[string]string) (int, []byte, error) {
	statusCode, body, err := httpGetHeader(requestURI, headers)
	PushLog(nil, username, name, requestURI, statusCode, body, err)
	return statusCode, body, err
}

func httpPost(requestBody []byte, requestURI string, headers map[string]string) (int, []byte, error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(requestURI)
	req.Header.SetMethod("POST")
	req.SetBody(requestBody)

	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	err := fc.DoTimeout(req, resp, apiTimeOut)
	return resp.StatusCode(), resp.Body(), err
}

func httpGet(requestURI string) (int, []byte, error) {

	return fc.GetTimeout(nil, requestURI, apiTimeOut)
}

func httpGetHeader(requestURI string, headers map[string]string) (int, []byte, error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	req.SetRequestURI(requestURI)
	req.Header.SetMethod("GET")
	//req.SetBody(requestBody)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	err := fc.DoTimeout(req, resp, apiTimeOut)

	return resp.StatusCode(), resp.Body(), err
}

func GetBalanceFromString(balance string) (string, error) {
	deciBalance, err := decimal.NewFromString(balance)
	if err != nil {
		return "", err
	}
	result := deciBalance.StringFixed(4)
	return result, nil
}

func GetBalanceFromByte(balance []byte) (string, error) {
	deciBalance, err := decimal.NewFromString(string(balance))
	if err != nil {
		return "", err
	}
	result := deciBalance.StringFixed(4)
	return result[:len(result)-2], nil
}

func GetBalanceFromFloat(balance float64) string {
	deciBalance := decimal.NewFromFloat(balance)
	result := deciBalance.StringFixed(4)
	return result
}

func News(socks5 string) {

	fc = &fasthttp.Client{
		MaxConnsPerHost: 60000,
		TLSConfig:       &tls.Config{InsecureSkipVerify: true},
		ReadTimeout:     apiTimeOut,
		WriteTimeout:    apiTimeOut,
	}

	if socks5 != "0.0.0.0" {
		fc.Dial = fasthttpproxy.FasthttpSocksDialer(socks5)
	}
}
