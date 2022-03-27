package party

import (
	//"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

// 统一验证各个接口参数

const (
	VNONE   int = 0
	VALNUM  int = 1
	VDIGIT  int = 2
	VALPHA  int = 3
	VAMOUNT int = 4
)

type callcheck func(val string, vmin, vmax int) bool

var validates = map[int]callcheck{
	VNONE:   CheckStringNone,
	VALNUM:  CheckStringAlnum,
	VDIGIT:  checkIntScope,
	VAMOUNT: checkFloat,
}

var rules = map[string]map[string]map[string]int{
	"confirm": map[string]map[string]int{
		"id": map[string]int{
			"rule": VDIGIT,
			"min":  1,
			"max":  9223372036854775807,
		},
	},
	"reg": map[string]map[string]int{
		"tester": map[string]int{
			"rule": VDIGIT,
			"min":  0,
			"max":  1,
		},
		"lang": map[string]int{
			"rule": VALNUM,
			"min":  1,
			"max":  3,
		},
		"username": map[string]int{
			"rule": VALNUM,
			"min":  3,
			"max":  12,
		},
		"password": map[string]int{
			"rule": VNONE,
			"min":  3,
			"max":  100,
		},
	},
	"login": map[string]map[string]int{
		"lang": map[string]int{
			"rule": VALNUM,
			"min":  1,
			"max":  3,
		},
		"deviceType": map[string]int{
			"rule": VDIGIT,
			"min":  1,
			"max":  4,
		},
		"username": map[string]int{
			"rule": VALNUM,
			"min":  3,
			"max":  12,
		},
		"password": map[string]int{
			"rule": VNONE,
			"min":  3,
			"max":  100,
		},
	},
	"balance": map[string]map[string]int{
		"username": map[string]int{
			"rule": VALNUM,
			"min":  3,
			"max":  12,
		},
		"password": map[string]int{
			"rule": VNONE,
			"min":  3,
			"max":  100,
		},
	},
	"transfer": map[string]map[string]int{
		"amount": map[string]int{
			"rule": VAMOUNT,
			"min":  0,
			"max":  2147483647,
		},
		"type": map[string]int{
			"rule": VALNUM,
			"min":  2,
			"max":  3,
		},
		"username": map[string]int{
			"rule": VALNUM,
			"min":  3,
			"max":  12,
		},
		"password": map[string]int{
			"rule": VNONE,
			"min":  3,
			"max":  100,
		},
	},
}

// 判断是否为float
func checkFloat(str string, vmin, vmax int) bool {

	val, err := decimal.NewFromString(str)
	if err != nil {
		return false
	}

	dmin := decimal.NewFromInt(int64(vmin))
	dmax := decimal.NewFromInt(int64(vmax))

	if val.Cmp(dmin) <= 0 || val.Cmp(dmax) == 1 {
		return false
	}

	return true
}

func CheckStringNone(s string, vmin, vmax int) bool {

	l := len(s)
	if l < 1 {
		return false
	}

	if l < vmin || l > vmax {
		return false
	}

	return true
}

func CheckStringAlnum(s string, vmin, vmax int) bool {

	l := len(s)
	if l < 1 {
		return false
	}

	if l < vmin || l > vmax {
		return false
	}

	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}

	return true
}

// 判断数字范围
func checkIntScope(s string, vmin, vmax int) bool {

	val, err := strconv.Atoi(s)
	if err != nil {
		return false
	}

	if val < vmin || val > vmax {
		return false
	}

	return true
}

func validator(rule map[string]map[string]int, param map[string]interface{}) (string, bool) {

	for name, values := range rule {

		val, ok := param[name].(string)
		if !ok {
			return name, false
		}

		key := values["rule"]

		if f, ok1 := validates[key]; ok1 {
			if ok2 := f(val, values["min"], values["max"]); !ok2 {
				return name, false
			}
		}
	}

	return "", true
}