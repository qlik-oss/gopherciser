package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

type (
	// Writer simple implementation of a file writer
	Writer struct {
		fil *os.File
	}
)

// NewWriter create new instance of Writer
func NewWriter(name string) (*Writer, error) {
	w := &Writer{}

	basePath := filepath.Dir(name)
	if basePath != "." {
		// "." is returned when there's no base path, hence no directory
		// needs to be created in advance
		if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "failed to create base folder <%s>", basePath)
		}
	}

	if err := w.createFile(name); err != nil {
		return nil, errors.Wrapf(err, "failed to create file<%s>", name)
	}
	return w, nil
}

// Write implement io.Writer interface
func (w Writer) Write(p []byte) (int, error) {
	if w.fil == nil {
		return 0, nil
	}

	n, err := w.fil.Write(p)
	if err != nil {
		return n, errors.WithStack(err)
	}

	return n, nil
}

func (w *Writer) createFile(name string) error {
	if fileExists(name) {
		name = backupName(name)
	}

	fil, err := os.Create(name)
	if err != nil {
		return errors.Wrapf(err, "Failed to create log file<%s>", name)
	}

	w.fil = fil
	return nil
}

// Close writer
func (w *Writer) Close() error {
	if w.fil == nil {
		return nil
	}

	return errors.WithStack(w.fil.Close())
}

func backupName(name string) string {
	if fileExists(name) {
		return backupName(addDashEnd(name))
	}
	return name
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return true
	}
	return false
}

func addDashEnd(name string) string {
	num := "-001"

	ext := filepath.Ext(name)
	runes := []rune(name[0:(len(name) - len(ext))])

	if len(name) > 4 && len(runes) > 3 && runes[len(runes)-4] == '-' {
		cNum := runes[len(runes)-3:]
		if cInt, err := strconv.Atoi(string(cNum)); err == nil { //else is not number, keep num
			runes = runes[0:(len(runes) - 4)] //Remove current dash numbering
			cInt++
			num = fmt.Sprintf("-%0.3d", cInt)
		}
	}

	return fmt.Sprintf("%s%s%s", string(runes), num, ext)
}
