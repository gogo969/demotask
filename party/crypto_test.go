package party

import (
	"net/url"
	"testing"
)

func TestTCGCP(t *testing.T) {

	// tcgcp test
	/*
		merchant_code=ybykylcny
		params=0PEVpOpWZosYOXBaEAQVg%2FdKyPlWD4MCZTCX6fZ%2BZ%2FsxbdbrKCDJWFASOmGZ6dh8DnylgKLoL4hPRdIYwJvd%2BGVyMJc30HuC5RIeFnxebHw%3D
		       0PEVpOpWZosYOXBaEAQVg/dKyPlWD4MCZTCX6fZ+Z/sxbdbrKCDJWFASOmGZ6dh8DnylgKLoL4hPRdIYwJvd+GVyMJc30HuC5RIeFnxebHw=
		sign=0ba9299f34c2d01fb0b98bbda03adde0ce7277e489740d5a33d695d0d3bd120f
	*/
	desKey := "P0LuJZGh"
	sha256Key := "PHCVVLcGZ8jJI4KS"
	s := `{"currency":"CNY","method":"cm","password":"1234567","username":"sar123456"}`
	e, err := desEncrypt([]byte(s), []byte(desKey))
	if err != nil {
		t.Error(err)
	}

	t.Log(e)

	t.Log(sha256sum([]byte(e + sha256Key)))

	param := url.Values{}
	param.Set("merchant_code", "ybykylcny")
	param.Set("params", e)
	param.Set("sign", sha256sum([]byte(string(e)+sha256Key)))

	t.Log(param.Encode())
}
