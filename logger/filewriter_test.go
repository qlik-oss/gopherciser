package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setTmpWorkDir changes the current working directory to the temp directory
// based on the OS. Hence the log files (and possibly folders) are created there
func setTmpWorkDir() error {
	if err := os.Chdir(os.TempDir()); err != nil {
		return err
	}
	return nil
}

func TestNewWriter(t *testing.T) {
	err := setTmpWorkDir()
	assert.NoError(t, err)
	file := "some_file"
	w, err := NewWriter(file)
	assert.NotNil(t, w)
	assert.NoError(t, err)
	_, statErr := os.Stat(file)
	assert.NoError(t, statErr)
	os.Remove(file)
}

func TestNewWriter_withFolder(t *testing.T) {
	err := setTmpWorkDir()
	assert.NoError(t, err)
	dir := "some_fancy_dir"
	file := "some_file"
	filePath := filepath.Join(dir, file)
	w, err := NewWriter(filePath)
	assert.NotNil(t, w)
	assert.NoError(t, err)
	_, statErr := os.Stat(filePath)
	assert.NoError(t, statErr)
	os.RemoveAll(dir)
}
