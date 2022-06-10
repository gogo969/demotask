package common

import (
	"task/contrib/apollo"
)

type Conf struct {
	Lang     string   `json:"lang"`
	Prefix   string   `json:"prefix"`
	Rocketmq []string `json:"rocketmq"`
	Db       struct {
		Master struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"master"`
		Report struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"report"`
		Bet struct {
			Addr        string `json:"addr"`
			MaxIdleConn int    `json:"max_idle_conn"`
			MaxOpenConn int    `json:"max_open_conn"`
		} `json:"bet"`
	} `json:"db"`
	Td struct {
		Addr        string `json:"addr"`
		MaxIdleConn int    `json:"max_idle_conn"`
		MaxOpenConn int    `json:"max_open_conn"`
	} `json:"td"`
	Redis struct {
		Addr     []string `json:"addr"`
		Password string   `json:"password"`
	} `json:"redis"`
}

func ConfParse(endpoints []string, path string) Conf {

	cfg := Conf{}

	apollo.New(endpoints)
	apollo.Parse(path, &cfg)
	apollo.Close()

	return cfg
}
