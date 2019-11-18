package main

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var errBreak = errors.New("break")

type walker struct {
	base     string
	hash     bool
	results  []*extInfo
	handlers []filepath.WalkFunc
}

func (w *walker) walk(path string, info os.FileInfo, ierr error) (err error) {
	logrus.WithFields(logrus.Fields{
		"path": path,
		"base": w.base,
		"name": info.Name(),
	}).Debug("walking")
	if path == w.base || path == "." {
		return nil
	}
	for _, h := range w.handlers {
		err = h(path, info, ierr)
		if err != nil {
			if err == errBreak {
				return nil
			}
			return err
		}
	}
	i, err := w.getInfo(path, info)
	if err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			return nil
		}
		return err
	}
	w.results = append(w.results, i)
	return err
}

func skipDirs(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return errBreak
	}
	return nil
}

func skipHidden(path string, info os.FileInfo, err error) error {
	if info.Name()[0] == '.' {
		if info.IsDir() {
			return filepath.SkipDir
		}
		return errBreak
	}
	return nil
}

func skipPermErr(path string, info os.FileInfo, err error) error {
	if err != nil {
		if os.IsPermission(err) {
			return nil
		}
		return err
	}
	return nil
}

type extInfo struct {
	os.FileInfo

	Path string
	MD5  string
}

func (i *extInfo) String() string {
	return i.Path
}

func (w *walker) getInfo(path string, info os.FileInfo) (*extInfo, error) {
	var (
		sum string
		err error
	)
	if w.hash {
		h := md5.New()
		if sum, err = hashFile(h, path); err != nil {
			return nil, err
		}
	}
	return &extInfo{
		FileInfo: info,
		Path:     path,
		MD5:      sum,
	}, nil
}

func hashFile(h hash.Hash, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}
	defer f.Close()

	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
