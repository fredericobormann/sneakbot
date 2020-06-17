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
}

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Group{})
}

func AddOrUpdateGroup(groupChatId int64, latestPollId int) {
	var group Group
	db.Where(Group{GroupchatId: groupChatId}).Assign(Group{LatestPollId: latestPollId}).FirstOrCreate(&group)
	fmt.Println(group)
}
