package cli

import (
	"github.com/codegangsta/cli"
	"fmt"
	"github.com/zanecloud/apiserver/utils"
	"os/exec"
	"github.com/Sirupsen/logrus"
)

const initCommandName = "init"

func initCommand(c *cli.Context) {

	//#准备加盐计算
	name := "root"
	salt := "1234567891234567"
	pass := "hell05a"
	content := fmt.Sprintf("%s:%s", pass, salt) //"$pass:$salt"
	//#生成加盐后的密码
	encryptedPassword := utils.Md5(content)

	cmd := "mongo"
	args := fmt.Sprintf("zanecloud --eval \"db.user.insertOne({name:'%s',pass:'%s',salt: '%s',roleset:4})\"", name, encryptedPassword, salt)
	_, err := exec.Command("sh", "-c", fmt.Sprintf("%s %s", cmd, args)).Output()
	if err != nil {
		logrus.Fatal(err)
	}

}
