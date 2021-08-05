package utils

import (
	"bufio"
	"io/ioutil"
	"os"
)

func IsExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func IsNotExists(path string) bool {
	return (!IsExists(path))
}

func FileGetContents(filename string) ([]byte, error) {
	fp, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)

	if err != nil {
		return nil, err
	}
	defer fp.Close()
	reader := bufio.NewReader(fp)
	contents, _ := ioutil.ReadAll(reader)

	return contents, nil
}

func FilePutContents(filename string, content []byte, opt int) error {
	fp, err := os.OpenFile(filename, opt, os.ModePerm)

	if err != nil {
		return err
	}

	defer fp.Close()

	_, err = fp.Write(content)

	return err
}

// MKDir 创建目录;传入绝对路径
func MKDir(dir string) (bool, error) {
	if ok := IsExists(dir); ok == true {
		return true, nil
	}
	err := os.MkdirAll(dir, os.ModePerm) //在当前目录下生成md目录
	if err != nil {
		return false, err
	}
	return true, nil
}
