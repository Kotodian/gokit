package upload

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Kotodian/gokit/file"
)

// FileUpload 文件上传
type FileUpload struct {
	file.FileBase

	RootPath string // 文件存取根目录
}

// New 创建一个上传文件对象
func New(rootpath string, fileSize ...int64) *FileUpload {
	uf := &FileUpload{
		RootPath: rootpath,
	}
	uf.SetAllowType()
	uf.SetMaxSize(0)
	if len(fileSize) == 1 {
		uf.SetMaxSize(fileSize[0])
	}
	return uf
}

// Upload 上传文件
// dirname	目录名，用以做目录区分。 可以传入 "station", "station/evse", "operator/logo" 等等
func (uf *FileUpload) Upload(r *http.Request, dirname string, randPath ...string) (fileurl string, err error) {
	if r.Method != "post" && r.Method != "POST" {
		return "", errors.New("method not allow")
	}

	if uf.RootPath == "" {
		return "", errors.New("rootpath is nil")
	}

	if uf.File, uf.FileHeader, err = r.FormFile("file"); err != nil {
		return "", err
	}
	defer uf.File.Close()

	if err = uf.IsValid(); err != nil {
		return "", err
	}

	var seed interface{}
	var dir, filename string
	if len(randPath) == 0 {
		rand.Seed(time.Now().UnixNano())
		seed = rand.Uint64()
		dir, filename = uf.GetHash(seed)
	} else {
		seed = randPath[0]
		dir, filename = uf.GetHash(randPath[0])
		filename = randPath[0]
	}

	filePath := fmt.Sprintf("static/ugc/images/%s/%s", dirname, dir)
	if err = uf.SaveTo(fmt.Sprintf("%s/%s", uf.RootPath, filePath), filename); err != nil {
		return "", err
	}

	fileurl = "/" + filePath + "/" + filename
	return fileurl, nil
}
