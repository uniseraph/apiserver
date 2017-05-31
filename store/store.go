package store

import (
	"gopkg.in/mgo.v2"
	"github.com/Sirupsen/logrus"
)

func SavePool(pool *PoolInfo) error {

	session, err := mgo.Dial("local")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB("zanecloud").C("pool")
	err = c.Insert(pool)
	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}