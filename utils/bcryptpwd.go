package utils

import (
	"database/sql"
	"github.com/jameskeane/bcrypt"
)

// bcrypt加密的密文
type BcryptPwd string

func NewBcryptPwd(ori string) BcryptPwd {
	salt, _ := bcrypt.Salt(10)
	pwd, _ := bcrypt.Hash(ori, salt)
	return BcryptPwd(pwd)
}

// for db scan
func (pwd *BcryptPwd) Scan(i interface{}) error {
	val := &sql.NullString{}
	if err := val.Scan(i); err != nil {
		return err
	}
	*pwd = BcryptPwd(val.String)
	return nil

}

// 原始数据与加密数据校验
func (pwd BcryptPwd) Match(ori string) bool {
	return bcrypt.Match(ori, string(pwd))
}

func (pwd BcryptPwd) String() string {
	return string(pwd)
}
