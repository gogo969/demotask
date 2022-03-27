package party

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"strings"
	"time"
)

var dgqpLang = map[string]string{
	"cn": "zh_cn", //简体中文
	"vn": "vn",    //越南
}

var dgqpCurrency = map[string]string{
	"cn": "1",  //简体中文
	"vn": "13", //越南
}

func dgqpPack(s, money string, param map[string]interface{}) (string, string) {

	account := param["prefix"].(string) + param["username"].(string)
	var orderid string
	args := map[string]string{
		"opcode": s,
		"acc":    account,
	}

	if s != "1" {
		ts := time.Now().Format("20160102150405999")
		args["score"] = money
		orderid = fmt.Sprintf("%s%s%s", param["agent_id"].(string), ts, account)
		args["order"] = orderid
	}

	if s == "0" {
		args["ip"] = param["ip"].(string)
		args["channelid"] = param["channelid"].(string)

		gCode := param["gamecode"].(string)
		if len(gCode) == 0 {
			gCode = "0"
		}

		args["gametype"] = gCode
	}

	var val []string
	for k, v := range args {
		val = append(val, fmt.Sprintf("%s=%s", k, v))
	}

	text := strings.Join(val, "&")
	encrypted := aesEcbEncrypt([]byte(text), []byte(param["des_key"].(string)))
	sEnc := b64.URLEncoding.EncodeToString(encrypted)
	sign := fmt.Sprintf("%s%s%s", param["agent_id"].(string), param["ms"].(string), param["md5_key"].(string))
	sign = MD5Hash(sign)
	url := fmt.Sprintf("%s?channel=%s&tm=%s&msg=%s&sign=%s", param["api"].(string), param["agent_id"].(string), param["ms"].(string), sEnc, sign)
	return url, orderid
}

//dgqpLogin
//0
func dgqpLogin(param map[string]interface{}) (int, string) {

	url, _ := dgqpPack("0", "0", param)
	statusCode, body, err := HttpGetHeaderWithLog(param["username"].(string), "DGQP", url, nil)
	if err != nil {
		return Failure, err.Error()
	}

	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	if v.GetInt("ret", "code") == 0 {
		return Success, string(v.GetStringBytes("ret", "url"))
	}

	return Failure, ""
}

//dgqpBalance
//1
func dgqpBalance(param map[string]interface{}) (int, string) {

	url, _ := dgqpPack("1", "0", param)
	statusCode, body, err := HttpGetWithPushLog(param["username"].(string), "DGQP", url)
	if err != nil {
		return Failure, err.Error()
	}

	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	if v.GetInt("ret", "code") == 0 {
		score := fmt.Sprintf("%f", v.GetFloat64("ret", "score"))
		amount, _ := GetBalanceFromByte([]byte(score))
		return Success, amount
	}

	return Failure, ""
}

//dgqpTransfer
//in 2
//out 3
func dgqpTransfer(param map[string]interface{}) (int, string) {

	flags := "3"
	if param["type"].(string) == "in" {
		flags = "2"
	}

	url, orderId := dgqpPack(flags, param["amount"].(string), param)
	statusCode, body, err := HttpGetWithPushLog(param["username"].(string), "DGQP", url)
	if err != nil {
		return Failure, err.Error()
	}

	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}

	if v.GetInt("code") == 0 {
		return Success, orderId
	}

	return Failure, ""
}
