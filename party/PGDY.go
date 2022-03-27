package party

import (
	"fmt"
	"net/url"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

const (
	PGDY = "PGDY"
)

var pgdyCurrency = map[string]string{
	"cn": "CNY",
	"vn": "VND",
}

var pgdyLang = map[string]string{
	"cn": "zh-CN",
	"vn": "VI",
}

func pgdyReg(param map[string]interface{}) (int, string) {

	var p fastjson.Parser
	args := map[string]string{
		"secret_key":     param["secret_key"].(string),
		"operator_token": param["operator_token"].(string),
		"player_name":    param["prefix"].(string) + param["username"].(string),
		"currency":       pgdyCurrency[param["lang"].(string)],
	}
	postData := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/Player/Create?%s", param["api"].(string), postData)
	u, _ := url.Parse(param["api"].(string))
	header := map[string]string{
		"Host": u.Host,
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), PGDY, reqUrl, header)
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

	vv := v.Get("data")

	if string(vv.GetStringBytes("action_result")) != "Success" {
		return Failure, string(vv.GetStringBytes("msg"))
	}

	verr := v.Get("error")
	if verr.String() == "null" || string(verr.GetStringBytes("code")) == "9411" {
		return Success, "success"
	}
	return Failure, string(vv.GetStringBytes("msg"))
}

func pgdyLogin(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	args := map[string]string{
		"secret_key":     param["secret_key"].(string),
		"operator_token": param["operator_token"].(string),
		"player_name":    param["prefix"].(string) + param["username"].(string),
		"game_code":      param["gamecode"].(string),
		"language":       pgdyLang[param["lang"].(string)],
	}
	postData := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/Launch?%s", param["api"].(string), postData)
	u, _ := url.Parse(param["api"].(string))
	header := map[string]string{
		"Host": u.Host,
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), PGDY, reqUrl, header)
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

	verr := v.Get("error")
	if verr.String() != "null" {
		return Failure, string(verr.GetStringBytes("message"))
	}

	vv := v.Get("data")

	return Success, string(vv.GetStringBytes("game_url"))
}

func pgdyBalance(param map[string]interface{}) (int, string) {

	var p fastjson.Parser
	args := map[string]string{
		"secret_key":     param["secret_key"].(string),
		"operator_token": param["operator_token"].(string),
		"player_name":    param["prefix"].(string) + param["username"].(string),
	}
	postData := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/GetPlayerBalance?%s", param["api"].(string), postData)
	u, _ := url.Parse(param["api"].(string))
	header := map[string]string{
		"Host": u.Host,
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), PGDY, reqUrl, header)
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

	verr := v.Get("error")

	if verr.String() != "null" {
		return Failure, string(verr.GetStringBytes("message"))
	}

	vv := v.Get("data")

	return Success, GetBalanceFromFloat(vv.GetFloat64("balance"))
}

func pgdyTransfer(param map[string]interface{}) (int, string) {

	var p fastjson.Parser
	args := map[string]string{
		"secret_key":     param["secret_key"].(string),
		"operator_token": param["operator_token"].(string),
		"player_name":    param["prefix"].(string) + param["username"].(string),
		"amount":         param["amount"].(string),
		"traceId":        param["id"].(string),
	}
	postData := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/TransferIn?%s", param["api"].(string), postData)
	if param["type"].(string) == "out" {
		reqUrl = fmt.Sprintf("%s/TransferOut?%s", param["api"].(string), postData)
	}
	u, _ := url.Parse(param["api"].(string))
	header := map[string]string{
		"Host": u.Host,
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), PGDY, reqUrl, header)
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

	vv := v.Get("error")
	if vv.String() == "null" {

		return Success, param["id"].(string)
	}

	return Failure, string(vv.GetStringBytes("message"))
}
