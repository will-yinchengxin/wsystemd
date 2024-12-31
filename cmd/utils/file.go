package utils

import (
	"os"
)

func GetFile(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
}

// WriteFile filepath = fmt.Sprintf("/tmp/proc-%s.pid", uid)
func WriteFile(filepath string, b []byte) error {
	return os.WriteFile(filepath, b, 0660)
}
