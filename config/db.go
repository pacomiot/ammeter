package config

import (
	"github.com/yuguorong/go/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type QdDB struct {
	dbSQL *gorm.DB
}

func (db *QdDB) ConnectQuarkSQL() {
	var err error

	db.dbSQL, err = gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func (db *QdDB) CloseQuarkSQL() {

}

func (db *QdDB) CreateTbl(cls interface{}) {
	err := db.dbSQL.AutoMigrate(cls)
	if err != nil {
		log.Error(err)
	}
}

func (db *QdDB) Save(tbl interface{}) {
	t := db.dbSQL.Save(tbl)
	if t.Error != nil {
		log.Info(t.Error)
	}
}

func (db *QdDB) Delete(tbl interface{}) {
	db.dbSQL.Where("1=1").Unscoped().Delete(tbl)
}

func (db *QdDB) First(tbl interface{}) {
	db.dbSQL.First(tbl)
}

func (db *QdDB) Last(tbl interface{}) {
	db.dbSQL.Last(tbl)
}

func (db *QdDB) Find(tbl interface{}, condition ...interface{}) {
	if len(condition) > 0 {
		db.dbSQL.Find(tbl, condition...)
	} else {
		db.dbSQL.Find(tbl)
	}
}

func (db *QdDB) List(tbl interface{}, field interface{}, name string) {
	db.dbSQL.Model(tbl).Association(name).Find(field)
}

func (db *QdDB) Gorm() *gorm.DB {
	return db.dbSQL
}

var Qdb *QdDB = nil

func GetDB() *QdDB {
	return Qdb
}

func initDB() (db *QdDB) {
	db = new(QdDB)
	db.ConnectQuarkSQL()
	return db
}

func init() {
	Qdb = initDB()
}
