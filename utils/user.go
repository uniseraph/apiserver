package utils

import (
	"github.com/zanecloud/apiserver/types"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
)

/*
	校验用户身份
	根据用户提供的密码，对比需要校验的用户实例
	通过对密码和盐做MD5运算，和user.Pass字段比较是否相同
*/
func ValidatePassword(user types.User, pass string) (ok bool, err error) {
	if (len(user.Pass) == 0) || (len(user.Salt) == 0) {
		ok = false
		err = errors.New("Password or Salt is empty!")
		return
	}

	ok = Md5(fmt.Sprintf("%s:%s", pass, user.Salt)) == user.Pass
	logrus.Debugf("getUserLogin::get the user %#v, password is %s, rlt:%d", user, pass, ok)
	return
}
