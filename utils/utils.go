package utils

import (
	"net/http"
	"github.com/Sirupsen/logrus"
)

//TODO using md5
func Md5(data string) string {
	return data
}


func HttpError(w http.ResponseWriter, err string, status int) {
	logrus.WithField("status", status).Errorf("HTTP error: %v", err)
	http.Error(w, err, status)
}
