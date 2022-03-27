package party

import (
	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
)

var (
	dialect   = g.Dialect("mysql")
	minAmount = decimal.NewFromFloat(1)
)

type callFunc func(map[string]interface{}) (int, string)

var routes = map[string]callFunc{

	//PG电子
	"pgdy_reg":      pgdyReg,
	"pgdy_login":    pgdyLogin,
	"pgdy_balance":  pgdyBalance,
	"pgdy_transfer": pgdyTransfer,

	//V8棋牌
	"kyqp_reg":      kyqpReg,
	"kyqp_login":    kyqpLogin,
	"kyqp_balance":  kyqpBalance,
	"kyqp_transfer": kyqpTransfer,

	//YB彩票
	"ybcp_reg":      ybcpReg,
	"ybcp_login":    ybcpLogin,
	"ybcp_balance":  ybcpBalance,
	"ybcp_transfer": ybcpTransfer,

	//DS
	"ds_reg":      dsReg,
	"ds_login":    dsLogin,
	"ds_balance":  dsBalance,
	"ds_transfer": dsTransfer,

	//dg
	"dgqp_reg":      dgqpLogin,
	"dgqp_login":    dgqpLogin,
	"dgqp_balance":  dgqpBalance,
	"dgqp_transfer": dgqpTransfer,
}
