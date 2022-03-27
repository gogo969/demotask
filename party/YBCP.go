package party

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"strings"
)

const YBCP string = "YBCP"

var ybcpLang = map[string]string{
	"cn": "zh",
	"vn": "vi",
}

var ybcpCurrency = map[string]string{
	"cn": "RMB",
	"vn": "VND",
}

func ybcpReg(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	accountType := "1"
	if param["tester"].(string) == "0" {
		accountType = "0"
	}

	args := map[string]string{
		"siteagent":   param["merchant_code"].(string),
		"loginname":   param["prefix"].(string) + param["username"].(string),
		"password":    param["password"].(string),
		"accounttype": accountType,
		//"Parentname":   "",
		"Lang":     ybcpLang[param["lang"].(string)],
		"Currency": ybcpCurrency[param["lang"].(string)],
		//"Categorytype": "",
	}

	args["sign"] = getMD5Hash(fmt.Sprintf("siteagent=%s&loginname=%s&password=%s&accounttype=%s%s",
		args["siteagent"],
		args["loginname"],
		args["password"],
		args["accounttype"],
		param["merchant_key"].(string)))

	str := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/api/CheckOrCreateGameAccount?%s", param["api"].(string), str)
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), YBCP, reqUrl, header)
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

	if v.GetBool("status") {
		return Success, fmt.Sprintf("%d", v.GetInt("code"))
	}

	return Failure, string(v.GetStringBytes("message"))
}

func ybcpLogin(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	sid := param["id"].(string) + param["ms"].(string)
	if len(sid) > 36 {
		sid = sid[0:36]
	}

	loginName := param["prefix"].(string) + param["username"].(string)
	if len(loginName) > 14 {
		loginName = loginName[0:14]
	}

	args := map[string]string{
		"siteagent": param["merchant_code"].(string),
		"loginname": loginName,
		"password":  param["password"].(string),
		"sid":       sid,
		"Lang":      ybcpLang[param["lang"].(string)],
	}

	sign := fmt.Sprintf("siteagent=%s&loginname=%s&password=%s&sid=%s",
		args["siteagent"],
		args["loginname"],
		args["password"],
		args["sid"])

	if len(param["gamecode"].(string)) > 0 {
		args["game"] = strings.ToUpper(param["gamecode"].(string))
		sign = fmt.Sprintf("%s&game=%s", sign, args["game"])
	}

	args["sign"] = getMD5Hash(sign + param["merchant_key"].(string))
	str := ParamEncode(args)

	reqUrl := fmt.Sprintf("%s/api/LoginGame?%s", param["api"].(string), str)
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), YBCP, reqUrl, header)
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

	if v.GetBool("status") {
		vv := v.Get("result")
		return Success, string(vv.GetStringBytes("action"))
	}

	return Failure, string(v.GetStringBytes("message"))
}

func ybcpBalance(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	loginName := param["prefix"].(string) + param["username"].(string)
	if len(loginName) > 14 {
		loginName = loginName[0:14]
	}

	args := map[string]string{
		"siteagent": param["merchant_code"].(string),
		"loginname": loginName,
		"password":  param["password"].(string),
	}

	args["sign"] = getMD5Hash(fmt.Sprintf("siteagent=%s&loginname=%s&password=%s%s",
		args["siteagent"],
		args["loginname"],
		args["password"],
		param["merchant_key"].(string)))

	str := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/api/GetBalance?%s", param["api"].(string), str)
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), YBCP, reqUrl, header)
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

	if v.GetBool("status") {
		vv := v.Get("result")
		return Success, GetBalanceFromFloat(vv.GetFloat64("money"))
	}

	return Failure, string(v.GetStringBytes("message"))
}

func ybcpTransfer(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	method := "IN"
	if param["type"] == "out" {
		method = "OUT"
	}

	loginName := param["prefix"].(string) + param["username"].(string)
	if len(loginName) > 14 {
		loginName = loginName[0:14]
	}

	id := param["id"].(string)
	if len(id) > 16 {
		id = id[len(id)-16:]
	}

	orderId := fmt.Sprintf("%s%s", param["merchant_code"].(string), id)

	args := map[string]string{
		"siteagent": param["merchant_code"].(string),
		"loginname": loginName,
		"password":  param["password"].(string),
		"billno":    orderId,
		"credit":    param["amount"].(string),
		"type":      method,
	}

	args["sign"] = getMD5Hash(fmt.Sprintf("siteagent=%s&loginname=%s&password=%s&billno=%s&credit=%s&type=%s%s",
		args["siteagent"],
		args["loginname"],
		args["password"],
		args["billno"],
		args["credit"],
		args["type"],
		param["merchant_key"].(string)))

	str := ParamEncode(args)
	reqUrl := fmt.Sprintf("%s/api/TransferCredit?%s", param["api"].(string), str)
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog(nil, param["username"].(string), YBCP, reqUrl, header)
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

	if v.GetBool("status") {
		return Success, orderId
	}

	return Failure, string(v.GetStringBytes("message"))
}
