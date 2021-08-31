package delete

import (
	"github.com/stretchr/testify/assert"
	"gorm.io/plugin/soft_delete"
	"testing"
	"time"
)

type UserWithSoftDelete struct {
	ID        uint                  `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Name      string                `gorm:"size:256;uniqueIndex:idx_name_deleted_at_unix"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:idx_name_deleted_at_unix"`
}

func TestUserWithSoftDelete(t *testing.T) {
	db := NewDB(t)

	//----------------------------------------------
	// Find function
	// > Query: SELECT * FROM `user_with_soft_deletes` WHERE `user_with_soft_deletes`.`deleted_at` = 0
	//----------------------------------------------
	var findUsers []*UserWithSoftDelete
	err := db.Find(&findUsers).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_soft_deletes` (`created_at`,`updated_at`,`name`,`deleted_at`) VALUES ('2021-08-22 01:52:01.415','2021-08-22 01:52:01.415','user1','0')
	//----------------------------------------------
	user1 := UserWithSoftDelete{Name: "user1"}
	err = db.Save(&user1).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// First function
	// > Query: SELECT * FROM `user_with_soft_deletes` WHERE `user_with_soft_deletes`.`id` = 1 AND `user_with_soft_deletes`.`deleted_at` = 0 ORDER BY `user_with_soft_deletes`.`id` LIMIT 1
	//----------------------------------------------
	var findUser UserWithSoftDelete
	err = db.First(&findUser, user1.ID).Error

	//----------------------------------------------
	// Delete function
	// > Query: UPDATE `user_with_soft_deletes` SET `deleted_at`=1629564721 WHERE `user_with_soft_deletes`.`id` = 1 AND `user_with_soft_deletes`.`deleted_at` = 0
	//----------------------------------------------
	err = db.Delete(&findUser).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_soft_deletes` (`created_at`,`updated_at`,`name`,`deleted_at`) VALUES ('2021-08-22 01:52:01.432','2021-08-22 01:52:01.432','user1','0')
	//----------------------------------------------
	user2 := UserWithSoftDelete{Name: user1.Name}
	err = db.Save(&user2).Error
	assert.NoError(t, err)
}
