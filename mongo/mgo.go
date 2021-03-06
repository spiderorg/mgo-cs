package mongo

import (
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/sirupsen/logrus"
	"github.com/spiderorg/mgo-cs/pool"
)

type MgoSrc struct {
	*mgo.Session
}

var (
	connGcSecond = time.Duration(100) * 1e9
	session      *mgo.Session
	err          error
	MgoPool      = pool.ClassicPool(
		1000,
		200,
		func() (pool.Src, error) {
			// if err != nil || session.Ping() != nil {
			// 	session, err = newSession()
			// }
			return &MgoSrc{session.Clone()}, err
		},
		connGcSecond)
)

// param:
// user : connection's username
// password : connection's password
// connect : connection's addr ----format like "127.0.0.1:27017/dbname"
// connCAP : max size of the pool
// gcSeconds : idl time of the per link
func Refresh(user, password, connect string, connCAP, gcSeconds int) {
	connGcSecond = time.Duration(gcSeconds) * 1e9

	MgoPool = pool.ClassicPool(
		connCAP,
		connCAP/5,
		func() (pool.Src, error) {
			// if err != nil || session.Ping() != nil {
			// 	session, err = newSession()
			// }
			return &MgoSrc{session.Clone()}, err
		},
		connGcSecond)

	url := fmt.Sprintf("mongodb://%s:%s@%s", user,
		password,
		connect,
	)

	session, err = mgo.Dial(url)
	if err != nil {
		logrus.Fatalln("MongoDB", err, "|", connect)
	} else if err = session.Ping(); err != nil {
		logrus.Fatalln("MongoDB", err, "|", connect)
	} else {
		session.SetPoolLimit(connCAP)
	}
	logrus.Infoln("To open mongo is ok.")
}

// 判断资源是否可用
func (self *MgoSrc) Usable() bool {
	if self.Session == nil || self.Session.Ping() != nil {
		return false
	}
	return true
}

// 使用后的重置方法
func (*MgoSrc) Reset() {}

// 被资源池删除前的自毁方法
func (self *MgoSrc) Close() {
	if self.Session == nil {
		return
	}
	self.Session.Close()
}

func Error() error {
	return err
}

// 调用资源池中的资源
func Call(fn func(pool.Src) error) error {
	return MgoPool.Call(fn)
}

// 销毁资源池
func Close() {
	MgoPool.Close()
}

// 返回当前资源数量
func Len() int {
	return MgoPool.Len()
}

// 获取所有数据
func DatabaseNames() (names []string, err error) {
	err = MgoPool.Call(func(src pool.Src) error {
		names, err = src.(*MgoSrc).DatabaseNames()
		return err
	})
	return
}

// 获取数据库集合列表
func CollectionNames(dbname string) (names []string, err error) {
	MgoPool.Call(func(src pool.Src) error {
		names, err = src.(*MgoSrc).DB(dbname).CollectionNames()
		return err
	})
	return
}
