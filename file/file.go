package file

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/Kotodian/gokit/utils"
)

var (
	// DefaultMaxSize 默认文件大小
	DefaultMaxSize int64 = 3145728

	// DefaultAllowType 默认支持格式
	DefaultAllowType = []string{"image/jpeg", "image/png"}
)

// File 文件处理
type FileBase struct {
	File          multipart.File
	FileHeader    *multipart.FileHeader
	AllowMimeType []string
	MaxSize       int64
}

// NewFile 创建
func NewFile(maxsize int64, allowType ...string) *FileBase {
	f := &FileBase{}
	f.SetMaxSize(maxsize)
	f.SetAllowType(allowType...)
	return f
}

// SetAllowType 设置文件类型
func (f *FileBase) SetAllowType(filetype ...string) {
	if len(filetype) > 0 {
		for k := range filetype {
			f.AllowMimeType = append(f.AllowMimeType, filetype[k])
		}
	} else {
		f.AllowMimeType = append(f.AllowMimeType, DefaultAllowType...)
	}

}

// SetMaxSize 设置最大文件
func (f *FileBase) SetMaxSize(maxsize int64) {
	f.MaxSize = DefaultMaxSize
	if maxsize != 0 {
		f.MaxSize = maxsize
	}
}

// IsValid 判断合法性
func (f *FileBase) IsValid() error {
	filetype := f.FileHeader.Header.Get("Content-Type")
	if filetype == "" {
		return errors.New("无法确定文件类型")
	}
	match := false
	for _, t := range f.AllowMimeType {
		if t == filetype {
			match = true
		}
	}
	if !match {
		return fmt.Errorf("暂不支持'%s'的文件类型", filetype)
	}
	if filesize := f.FileHeader.Size; filesize > f.MaxSize {
		return fmt.Errorf("文件大小超出了限制")
	}
	return nil
}

// GetHash 生成相对路径和文件名
func (f *FileBase) GetHash(param interface{}) (string, string) {
	filename := fmt.Sprintf("%v_%s", param, f.FileHeader.Filename)
	dir := utils.MD5(filename)
	return fmt.Sprintf("%s/%s/%s", dir[0:2], dir[2:4], dir[4:6]), filename
}

// SaveTo 保存文件到指定路径
func (f *FileBase) SaveTo(dir, filename string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	out, err := os.Create(fmt.Sprintf("%s/%s", dir, filename))
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, f.File); err != nil {
		return err
	}
	return nil
}
