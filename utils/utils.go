package utils

import (
	"github.com/Sirupsen/logrus"
	"net/http"
	"crypto/rand"
	"crypto/md5"
	"io"
	"fmt"
)

//计算MD5
func Md5(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	str := fmt.Sprintf("%x", h.Sum(nil))
	logrus.Debugf("MD5 for string: %s, hash is %s", data, str)
	return str
}

func HttpError(w http.ResponseWriter, err string, status int) {
	logrus.WithField("status", status).Errorf("HTTP error: %v", err)
	http.Error(w, err, status)
}

//生成随机字符串，长度为n
func RandomStr(n int) string {
	if n > 0 {
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		s := fmt.Sprintf("%X", b)

		return s
	}
	return ""
}