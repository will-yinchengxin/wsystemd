package params

type JobRun struct {
	Type    string   `json:"type" validate:"omitempty"`
	Cmd     string   `json:"cmd" validate:"omitempty"`
	Args    []string `json:"args" validate:"omitempty"`
	Outfile string   `json:"outfile" validate:"required,min=1"`
	Errfile string   `json:"errfile" validate:"required,min=1"`
}

type JobCheck struct {
	Type     string   `json:"type" validate:"omitempty"`
	Api      string   `json:"api" validate:"omitempty"`
	Cmd      string   `json:"cmd" validate:"omitempty"`
	Args     []string `json:"args" validate:"omitempty"`
	Interval int      `json:"interval" validate:"omitempty"`
	FailNum  int      `json:"failNum" validate:"omitempty"`
}

type JobCfg struct {
	Type      string     `json:"type" validate:"omitempty"`
	Num       int        `json:"num" validate:"omitempty"`
	Checks    []JobCheck `json:"checks" validate:"omitempty"`
	FailCodes []int      `json:"failCodes" validate:"omitempty"`

	// 新版本需要的参数
	DoOnce     bool   `json:"doOnce" validate:"omitempty"`
	Run        JobRun `json:"run" validate:"required"`
	Restart    string `json:"restart" validate:"omitempty"`
	Node       string `json:"node" validate:"omitempty"`
	Dc         string `json:"dc" validate:"omitempty"`
	Ip         string `json:"ip" validate:"omitempty"`
	LoadMethod string `json:"loadMethod" validate:"omitempty"`

	BigOne      string `json:"bigOne" validate:"omitempty"`
	BigOneJobId string `json:"bigOneJobId" validate:"omitempty"`
}

type JobReporter struct {
	Token string `form:"token" validate:"required"`
	Pid   string `form:"pid" validate:"required"`
}

type BigOne struct {
	BigOneJobId string `json:"bigOneJobId" validate:"required"`
}
