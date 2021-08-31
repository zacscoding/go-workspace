package delete

import (
	"github.com/stretchr/testify/assert"
	"gorm.io/plugin/soft_delete"
	"testing"
	"time"
)

type UserWithDeleteFlag struct {
	ID        uint                  `json:"id" gorm:"primarykey"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
	Name      string                `gorm:"size:256;unique"`
	IsDeleted soft_delete.DeletedAt `gorm:"softDelete:flag"`
}

func TestUserWithDeleteFlag(t *testing.T) {
	db := NewDB(t)

	//----------------------------------------------
	// Find function
	// > Query: SELECT * FROM `user_with_delete_flags` WHERE `user_with_delete_flags`.`is_deleted` = 0
	//----------------------------------------------
	var findUsers []*UserWithDeleteFlag
	err := db.Find(&findUsers).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// Save function
	// > Query: INSERT INTO `user_with_delete_flags` (`created_at`,`updated_at`,`name`,`is_deleted`) VALUES ('2021-08-22 01:55:11.175','2021-08-22 01:55:11.175','user1','0')
	//----------------------------------------------
	user1 := UserWithDeleteFlag{Name: "user1"}
	err = db.Save(&user1).Error
	assert.NoError(t, err)

	//----------------------------------------------
	// First function
	// > Query: SELECT * FROM `user_with_soft_deletes` WHERE `user_with_soft_deletes`.`id` = 1 AND `user_with_soft_deletes`.`deleted_at` = 0 ORDER BY `user_with_soft_deletes`.`id` LIMIT 1
	//----------------------------------------------
	var findUser UserWithDeleteFlag
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
	user2 := UserWithDeleteFlag{Name: user1.Name}
	err = db.Save(&user2).Error
	assert.NoError(t, err)
}
