package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

type Group struct {
	gorm.Model
	GroupchatId  int64
	LatestPollId int
	Activated    *bool `gorm:"default:true"`
}

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Group{})
}

func AddOrUpdateGroup(groupChatId int64, latestPollId int) {
	var group Group
	t := true
	db.Where(Group{GroupchatId: groupChatId}).Assign(Group{LatestPollId: latestPollId, Activated: &t}).FirstOrCreate(&group)
	fmt.Println(group)
}

func DeactivateGroup(groupChatId int64) {
	f := false
	var group Group
	db.Where(Group{GroupchatId: groupChatId}).First(&group)
	if group.Activated != nil {
		group.Activated = &f
		db.Save(&group)
	}
}
