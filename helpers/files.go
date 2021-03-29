package helpers

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func AppendToBaseName(filePath string, strs ...string) string {
	ext := path.Ext(filePath)
	return strings.TrimSuffix(filePath, ext) + strings.Join(strs, "") + ext
}

// WriteToFile writes data to file defined by filePath
func WriteToFile(filePath string, data []byte) error {
	dir := path.Dir(filePath)
	if len(dir) > 0 {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.Errorf("could not create directory<%s>: %v", dir, err)
		}
	}
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return errors.Errorf("could not write to file<%s>: %v", filePath, err)
	}
	return nil
}

var unsafeFileNameChar = regexp.MustCompile(`[<>:"/\\|?*]`)

func ToValidWindowsFileName(fileName string) string {
	return unsafeFileNameChar.ReplaceAllLiteralString(fileName, "_")
}
