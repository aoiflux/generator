package util

import (
	"crypto/rand"
	"encoding/base32"
	"generator/constant"
	"os"
	"path/filepath"
)

func GetAbsPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	finfo, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if finfo.IsDir() {
		return absPath, nil
	}

	return filepath.Dir(absPath), nil
}

func GetRandomString(length int) string {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return base32.StdEncoding.EncodeToString(randomBytes)[:length]
}

func GetFilePath(basedir string, ext string) string {
	filename := GetRandomString(constant.FileNameLen) + ext
	return filepath.Join(basedir, filename)
}
