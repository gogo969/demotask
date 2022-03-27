package party

import (
	b64 "encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const KYQP string = "KYQP"

func kyqpPack(desStr string, param map[string]interface{}) string {

	md5Str := fmt.Sprintf("%s%s%s", param["agent"].(string), param["ms"].(string), param["md5_key"].(string))

	aesByte := aesEcbEncrypt([]byte(desStr), []byte(param["des_key"].(string)))

	args := url.Values{}
	args.Set("param", b64.StdEncoding.EncodeToString(aesByte))
	args.Set("key", getMD5Hash(md5Str))

	return args.Encode()
}

// 注册
func kyqpReg(param map[string]interface{}) (int, string) {

	return kyqpLogin(param)
}

// 登陆
func kyqpLogin(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	username := param["prefix"].(string) + param["username"].(string)
	orderid := fmt.Sprintf("%s%s%s", param["agent"].(string), time.Now().Format("20060102150405000"), username)
	signStr := fmt.Sprintf("s=0&account=%s&money=0&orderid=%s&ip=%s&lineCode=%s&KindID=0", username, orderid, param["ip"].(string), param["prefix"].(string))
	paramStr := kyqpPack(signStr, param)
	requestURI := fmt.Sprintf("%s/channelHandle?agent=%s&timestamp=%s&%s", param["api"].(string), param["agent"].(string), param["ms"].(string), paramStr)

	statusCode, body, err := HttpGetWithPushLog(param["username"].(string), KYQP, requestURI)
	if err != nil {
		return Failure, err.Error()
	}

	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	d := v.Get("d")
	if d.GetInt("code") == 0 {
		return Success, string(d.GetStringBytes("url"))
	}

	return Failure, fmt.Sprintf("%d", d.GetInt("code"))
}

// 查询余额
func kyqpBalance(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	username := param["prefix"].(string) + param["username"].(string)
	signStr := fmt.Sprintf("s=1&account=%s", username)
	paramStr := kyqpPack(signStr, param)
	requestURI := fmt.Sprintf("%s/channelHandle?agent=%s&timestamp=%s&%s", param["api"].(string), param["agent"].(string), param["ms"].(string), paramStr)

	statusCode, body, err := HttpGetWithPushLog(param["username"].(string), KYQP, requestURI)
	if err != nil {
		return Failure, err.Error()
	}

	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	d := v.Get("d")
	if d.GetInt("code") == 0 {
		return Success, GetBalanceFromFloat(d.GetFloat64("money"))
	}

	return Failure, fmt.Sprintf("%d code error", d.GetInt("code"))
}

// 上下分
func kyqpTransfer(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	username := param["prefix"].(string) + param["username"].(string)
	orderid := fmt.Sprintf("%s%s%s", param["agent"].(string), time.Now().Format("20060102150405000"), username)
	s := "2"
	if param["type"].(string) == "out" {
		s = "3"
	}
	signStr := fmt.Sprintf("s=%s&account=%s&money=%s&orderid=%s", s, username, param["amount"].(string), orderid)
	paramStr := kyqpPack(signStr, param)
	requestURI := fmt.Sprintf("%s/channelHandle?agent=%s&timestamp=%s&%s", param["api"].(string), param["agent"].(string), param["ms"].(string), paramStr)
	fmt.Println(requestURI)
	statusCode, body, err := HttpGetWithPushLog(param["username"].(string), KYQP, requestURI)
	if err != nil {
		return Failure, err.Error()
	}
	fmt.Println(string(body))
	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	d := v.Get("d")
	if d.GetInt("code") == 0 {
		return Success, orderid
	}

	return Failure, fmt.Sprintf("%d", d.GetInt("code"))
}
