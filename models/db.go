package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func init() {
	initDB()
}

var (
	o orm.Ormer
)

func initDB() {
	orm.Debug = true
	if err := orm.RegisterDriver("sqlite", orm.DRSqlite); err != nil {
		log.Fatalf("orm.RegisterDrive fail: %+v", err)
	}
	if err := orm.RegisterDataBase("default", "sqlite3", "pms.db"); err != nil {
		log.Fatalf("orm.RegisterDataBase fail: %+v", err)
	}

	orm.RegisterModelWithPrefix("pms_", new(User), new(Task))
	if err := orm.RunSyncdb("default", false, true); err != nil {
		log.Fatalf("orm.RunSyncdb fail: %+v", err)
	}

	o = orm.NewOrm()
	if err := o.Using("default"); err != nil {
		log.Fatalf("orm.Using database fail: %+v", err)
	}

	if err := initUser(); err != nil {
		log.Printf("initUser fail: %+v", err)
	}
}

func initUser() error {
	if created, id, err := o.ReadOrCreate(&User{Name: "admin", Password: "admin"}, "name"); err != nil {
		if created {
			log.Printf("new User, id: %d", id)
		}
	}

	return nil
}