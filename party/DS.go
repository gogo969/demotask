package party

import (
	"crypto/md5"
	"crypto/rand"
	b64 "encoding/base64"
	"fmt"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"github.com/wI2L/jettison"
	"io"
)

const DS string = "DS"

var dsLang = map[string]string{
	"cn": "zh_cn",
	"vn": "vi_vn",
}

var dsCurrency = map[string]string{
	"cn": "CNY",
	"vn": "1VND",
}

func dsPack(channel, aesKey, signKey string, args map[string]interface{}) (string, error) {

	data, _ := jettison.Marshal(args)

	salt := make([]byte, 8)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	salted := ""
	dI := ""

	for len(salted) < 48 {
		md := md5.New()
		md.Write([]byte(dI + aesKey + string(salt)))
		dM := md.Sum(nil)
		dI = string(dM[:16])
		salted = salted + dI
	}

	data = aesCbcEncrypt(data, salted[0:32], salted[32:48])
	body := b64.StdEncoding.EncodeToString([]byte("Salted__" + string(salt) + string(data)))

	return fmt.Sprintf(`{"channel":"%s","data":"%s","sign":"%s"}`, channel, body, getMD5Hash(fmt.Sprintf("%s%s", body, signKey))), nil
}

func dsReg(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	args := map[string]interface{}{
		"agent":    param["agent"].(string),
		"account":  param["prefix"].(string) + param["username"].(string),
		"password": param["password"].(string),
	}

	str, err := dsPack(param["channel"].(string), param["aes_key"].(string), param["sign_key"].(string), args)
	if err != nil {
		return Failure, err.Error()
	}
	reqUrl := fmt.Sprintf("%s/v1/member/create", param["api"].(string))
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog([]byte(str), param["username"].(string), DS, reqUrl, header)
	if err != nil {
		return Failure, err.Error()
	}
	fmt.Println("=================ds reg===================")
	fmt.Println(string(body))
	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}
	vv := v.Get("result")
	if vv.GetInt("code") == 1 || vv.GetInt("code") == 2 {
		return Success, "success"
	}

	return Failure, string(vv.GetStringBytes("msg"))
}

func dsLogin(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	args := map[string]interface{}{
		"game_id": param["gamecode"].(string),
		"agent":   param["agent"].(string),
		"account": param["prefix"].(string) + param["username"].(string),
		"lang":    dsLang[param["lang"].(string)],
	}

	str, err := dsPack(param["channel"].(string), param["aes_key"].(string), param["sign_key"].(string), args)
	if err != nil {
		return Failure, err.Error()
	}

	reqUrl := fmt.Sprintf("%s/v1/member/login_game", param["api"].(string))
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog([]byte(str), param["username"].(string), DS, reqUrl, header)
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
	vv := v.Get("result")
	if vv.GetInt("code") == 1 {
		return Success, string(v.GetStringBytes("url"))
	}

	return Failure, string(vv.GetStringBytes("msg"))
}

func dsBalance(param map[string]interface{}) (int, string) {

	var p fastjson.Parser
	args := map[string]interface{}{
		"agent":   param["agent"].(string),
		"account": param["prefix"].(string) + param["username"].(string),
	}

	str, err := dsPack(param["channel"].(string), param["aes_key"].(string), param["sign_key"].(string), args)
	if err != nil {
		return Failure, err.Error()
	}

	reqUrl := fmt.Sprintf("%s/v1/trans/check_balance", param["api"].(string))
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog([]byte(str), param["username"].(string), DS, reqUrl, header)
	if err != nil {
		return Failure, err.Error()
	}
	fmt.Println("=================ds balance===================")
	fmt.Println(string(body))
	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}
	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}
	vv := v.Get("result")
	if vv.GetInt("code") == 1 {
		balance, _ := GetBalanceFromByte(v.GetStringBytes("balance"))
		return Success, balance
	}

	return Failure, string(vv.GetStringBytes("msg"))
}

func dsTransfer(param map[string]interface{}) (int, string) {

	var p fastjson.Parser

	method := 1
	if param["type"] == "out" {
		method = 0
	}

	args := map[string]interface{}{
		"serial":    param["id"].(string),
		"agent":     param["agent"].(string),
		"account":   param["prefix"].(string) + param["username"].(string),
		"amount":    param["amount"].(string),
		"oper_type": method,
	}

	str, err := dsPack(param["channel"].(string), param["aes_key"].(string), param["sign_key"].(string), args)
	if err != nil {
		return Failure, err.Error()
	}

	reqUrl := fmt.Sprintf("%s/v1/trans/transfer", param["api"].(string))
	header := map[string]string{
		"Content-Type": "application/json;charset=utf-8",
	}

	statusCode, body, err := HttpPostWithPushLog([]byte(str), param["username"].(string), DS, reqUrl, header)
	if err != nil {
		return Failure, err.Error()
	}
	fmt.Println("=================ds transfer===================")
	fmt.Println(string(body))
	if statusCode != fasthttp.StatusOK {
		return Failure, fmt.Sprintf("%d", statusCode)
	}

	v, err := p.ParseBytes(body)
	if err != nil {
		return Failure, err.Error()
	}
	vv := v.Get("result")
	if vv.GetInt("code") == 1 {
		return Success, string(v.GetStringBytes("trans_id"))
	}

	return Failure, string(vv.GetStringBytes("msg"))
}
