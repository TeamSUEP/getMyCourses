package config

import (
	"net/http"
)

const SupwisdomUrl = "https://jw.shiep.edu.cn"
const IdsUrl = "https://ids.shiep.edu.cn"
const IdsService = "http://jw.shiep.edu.cn/eams/login.action"

const UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:108.0) Gecko/20100101 Firefox/108.0"

var Tr = &http.Transport{
	// TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	Proxy: http.ProxyFromEnvironment,
}
