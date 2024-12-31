package core

import (
	"os"
	"wsystemd/cmd/http/consts"

	"github.com/spf13/viper"
)

var (
	CoreConfig   map[string]interface{}
	ConfigInited bool
)

type Setting struct {
	vp *viper.Viper
}

func newSetting() (*Setting, error) {
	vp := viper.New()
	if pathExists(consts.ConfigPath) {
		vp.AddConfigPath(consts.ConfigPath)
	} else {
		vp.AddConfigPath(consts.ConfigLocalPath)
	}
	vp.SetConfigName(consts.ConfigName)
	vp.SetConfigType(consts.ConfigType)
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return &Setting{vp}, nil
}

func (s *Setting) readSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func FetchCoreConfig() error {
	if ConfigInited {
		return nil
	}
	setting, err := newSetting()
	if err != nil {
		return err
	}
	CoreConfig = make(map[string]interface{})
	err = setting.readSection(consts.ConfigName, &CoreConfig)
	if err != nil {
		return err
	}
	ConfigInited = true
	return nil
}
