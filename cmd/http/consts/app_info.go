package consts

const (
	ConfigPath      = "/usr/local/vs_conf"
	ConfigLocalPath = "conf"
	ConfigName      = "wsystemd"
	ConfigType      = "yaml"
)

const (
	NotDoOnce = iota + 1
	DoOnce
)

const (
	Load_Method_RR   = "round_robin"
	Load_Method_HASH = "hash"
	Load_Method_CPU  = "cpu"
	Load_Method_LOAD = "load"

	BigOne = "bigOne"
)

const (
	ServerPort = "9900"
)
