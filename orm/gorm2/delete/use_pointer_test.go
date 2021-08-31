package delete

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type UserWithPointer struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"size:256;uniqueIndex:idx_name_active"`
	Active    *bool     `gorm:"uniqueIndex:idx_name_active;default:1"`
}

func TestUserWithPointer(t *testing.T) {
	db := NewDB(t)

	//----------------------------------------------
	// Find function
	// > Query: SELECT * FROM `user_with_pointers`
	//----------------------------------------------
	var findUsers []*UserWithPointer
	err := db.Find(&findUsers).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Find function
	// > Query: SELECT * FROM `user_with_pointers` WHERE active IS NOT NULL
	//----------------------------------------------
	err = db.Where("active IS NOT NULL").Find(&findUsers).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_pointers` (`created_at`,`updated_at`,`name`,`active`) VALUES ('2021-08-22 02:10:08.069','2021-08-22 02:10:08.069','user1',true)
	//----------------------------------------------
	user1 := UserWithPointer{Name: "user1"}
	err = db.Save(&user1).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// First function
	// > Query: SELECT * FROM `user_with_pointers` WHERE `user_with_pointers`.`id` = 1 ORDER BY `user_with_pointers`.`id` LIMIT 1
	//----------------------------------------------
	var findUser UserWithPointer
	err = db.First(&findUser, user1.ID).Error

	//----------------------------------------------
	// Delete function
	// > Query: UPDATE `user_with_pointers` SET `created_at`='2021-08-22 02:16:55.904',`updated_at`='2021-08-22 02:16:55.912',`name`='user1',`active`=NULL WHERE `id` = 1
	//----------------------------------------------
	findUser.Active = nil
	err = db.Save(&findUser).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_pointers` (`created_at`,`updated_at`,`name`,`active`) VALUES ('2021-08-22 02:25:33.825','2021-08-22 02:25:33.825','user1',true)
	//----------------------------------------------
	user2 := UserWithPointer{Name: user1.Name}
	err = db.Save(&user2).Error
	assert.NoError(t, err)
}
