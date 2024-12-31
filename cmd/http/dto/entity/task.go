package entity

import "time"

type Task struct {
	ID            int64     `gorm:"column:id" json:"id" form:"id"`
	JobId         string    `gorm:"column:job_id" json:"job_id" form:"job_id"`
	Node          string    `gorm:"column:node" json:"node" form:"node"`
	Pid           int       `gorm:"column:pid" json:"pid" form:"pid"`
	Cmd           string    `gorm:"column:cmd" json:"cmd" form:"cmd"`
	Args          string    `gorm:"column:args" json:"args" form:"args"`
	Outfile       string    `gorm:"column:outfile" json:"outfile" form:"outfile"`
	Errfile       string    `gorm:"column:errfile" json:"errfile" form:"errfile"`
	Dc            string    `gorm:"column:dc" json:"dc" form:"dc"`
	Ip            string    `gorm:"column:ip" json:"ip" form:"ip"`
	BigOne        string    `gorm:"column:big_one" json:"big_one" form:"big_one"`
	Type          string    `gorm:"column:type" json:"type" form:"type"`
	LoadMethod    string    `gorm:"column:load_method" json:"load_method" form:"load_method"`
	DoOnce        int64     `gorm:"column:do_once" json:"do_once" form:"do_once"`
	Status        int64     `gorm:"column:status" json:"status" form:"status"`
	RetryCount    int64     `gorm:"column:retry_count" json:"retry_count" form:"retry_count"`
	LastError     string    `gorm:"column:last_error" json:"last_error" form:"last_error"`
	HeartBeatTime time.Time `gorm:"column:heart_beat_time" json:"heart_beat_time" form:"heart_beat_time"`
	CreateTime    time.Time `gorm:"column:create_time" json:"create_time" form:"create_time"`
	UpdateTime    time.Time `gorm:"column:update_time" json:"update_time" form:"update_time"`
}

func (t *Task) TableName() string {
	return "task"
}
