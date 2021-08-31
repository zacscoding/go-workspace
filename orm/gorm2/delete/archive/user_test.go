package archive

import (
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"
)

func TestEmbeddedUser(t *testing.T) {
	db := NewDB(t)
	getUser := func(id uint) *EmbeddedUser {
		var find EmbeddedUser
		if err := db.Unscoped().First(&find, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			t.Fatalf("find user:%v", err)
		}
		return &find
	}

	// (1) save
	u := EmbeddedUser{
		Name: "user1",
	}
	// INSERT INTO `embedded_users` (`created_at`,`updated_at`,`deleted_at`,`name`) VALUES ("2020-10-11 19:45:54.869","2020-10-11 19:45:54.869",NULL,"user1")
	if err := db.Save(&u).Error; err != nil {
		t.Fatalf("save user:%v", err)
	}
	fmt.Println(">> After create:", getUser(u.ID).String())
	time.Sleep(3 * time.Second)

	// (2) updated
	updateUser := getUser(u.ID)
	updateUser.Name = "update-user1"
	// UPDATE `embedded_users` SET `created_at`="2020-10-11 19:45:54.87",`updated_at`="2020-10-11 19:45:57.881",`deleted_at`=NULL,`name`="update-user1" WHERE `id` = 1
	if err := db.Save(updateUser).Error; err != nil {
		t.Fatalf("update user:%v", err)
	}
	fmt.Println(">> After update:", getUser(u.ID).String())
	time.Sleep(3 * time.Second)

	// (3) delete
	// UPDATE `embedded_users` SET `deleted_at`="2020-10-11 19:46:00.894" WHERE `embedded_users`.`id` = 1
	if err := db.Delete(&updateUser).Error; err != nil {
		t.Fatalf("delete user:%v", err)
	}
	fmt.Println(">> After delete:", getUser(u.ID).String())

	// Output
	// >> After create: EmbeddedUser{ID:1, CreatedAt:2020-10-11 19:42:07.104 +0900 KST, UpdatedAt:2020-10-11 19:42:07.104 +0900 KST, DeletedAt:{0001-01-01 00:00:00 +0000 UTC false}, Name:user1}
	// >> After update: EmbeddedUser{ID:1, CreatedAt:2020-10-11 19:42:07.104 +0900 KST, UpdatedAt:2020-10-11 19:42:10.119 +0900 KST, DeletedAt:{0001-01-01 00:00:00 +0000 UTC false}, Name:update-user1}
	// >> After delete: EmbeddedUser{ID:1, CreatedAt:2020-10-11 19:42:07.104 +0900 KST, UpdatedAt:2020-10-11 19:42:10.119 +0900 KST, DeletedAt:{2020-10-11 19:42:13.133 +0900 KST true}, Name:update-user1}
}

func TestEmbeddedUser_DeleteAndSaveAgain(t *testing.T) {
	db := NewDB(t)

	u := EmbeddedUser{
		Name: "user1",
	}
	if err := db.Save(&u).Error; err != nil {
		t.Fatalf("save an user:%v", err)
	}
	if err := db.Delete(&u).Error; err != nil {
		t.Fatalf("delete an user:%v", err)
	}

	newUser := EmbeddedUser{
		Name: u.Name,
	}

	err := db.Save(&newUser).Error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	fmt.Printf("save err:%v\n", err)
	// Output
	// save err:Error 1062: Duplicate entry 'user1' for key 'name'
}

func TestSolution1_User(t *testing.T) {
	db := NewDB(t)
	getUser := func(id uint) *Sol1User {
		var (
			find Sol1User
		)
		if err := db.First(&find, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			t.Fatalf("find user:%v", err)
		}
		return &find
	}

	// save
	u := Sol1User{
		Name: "user1",
	}
	if err := db.Save(&u).Error; err != nil {
		t.Fatalf("save an user:%v", err)
	}
	find := getUser(u.ID)
	fmt.Println(">> After save:", find.String())
	// delete
	find.DeletedAtUnix = time.Now().Unix()
	if err := db.Save(&find).Error; err != nil {
		t.Fatalf("delete an user:%v", err)
	}
	find = getUser(u.ID)
	fmt.Println(">> After delete:", find.String())
	// Output
	// >> After save: Sol1User{ID:1, CreatedAt:2020-10-11 20:40:34.855 +0900 KST, UpdatedAt:2020-10-11 20:40:34.855 +0900 KST, DeletedAtUnix:0, Name:user1}
	// >> After delete: Sol1User{ID:1, CreatedAt:2020-10-11 20:40:34.855 +0900 KST, UpdatedAt:2020-10-11 20:40:34.862 +0900 KST, DeletedAtUnix:1602416434, Name:user1}
}

func TestSolution2_User(t *testing.T) {
	db := NewDB(t)

	getUser := func(id uint) *Sol2User {
		var (
			find Sol2User
		)
		if err := db.First(&find, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			t.Fatalf("find user:%v", err)
		}
		return &find
	}

	// 1) save
	u := Sol2User{
		Name: "user1",
	}
	if err := db.Save(&u).Error; err != nil {
		t.Fatalf("save an user:%v", err)
	}

	u12 := Sol2User{
		Name: u.Name,
	}
	if err := db.Save(&u12).Error; err != nil {
		fmt.Printf("Save an user:%v\n", err)
	}
	// 2) delete
	find := getUser(u.ID)
	find.Active = nil
	if err := db.Save(&find).Error; err != nil {
		t.Fatalf("delete an user:%v", err)
	}
	find = getUser(u.ID)
	fmt.Println(">> After delete:", find.String())

	// 3) save
	u2 := Sol2User{
		Name: u.Name,
	}
	if err := db.Save(&u2).Error; err != nil {
		t.Fatalf("save an user:%v", err)
	}
	// 4) delete
	find = getUser(u2.ID)
	find.Active = nil
	if err := db.Save(&find).Error; err != nil {
		t.Fatalf("delete an user:%v", err)
	}
	find = getUser(u2.ID)
	fmt.Println(">> After delete:", find.String())
}

func TestSolution3_User(t *testing.T) {
	db := NewDB(t)
	getUser := func(id uint) *Sol3User {
		var (
			find Sol3User
		)
		if err := db.First(&find, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			t.Fatalf("find user:%v", err)
		}
		return &find
	}

	var findUsers []*Sol3User
	if err := db.Find(&findUsers).Error; err != nil {
		t.Fatalf("find users: %v", err)
	}

	// save
	u := Sol3User{
		Name: "user1",
	}
	if err := db.Save(&u).Error; err != nil {
		t.Fatalf("save an user:%v", err)
	}
	find := getUser(u.ID)
	fmt.Println(">> After save:", find.String())

	// delete
	if err := db.Delete(find).Error; err != nil {
		t.Fatalf("delete an user:%v", err)
	}
	find = getUser(u.ID)
	fmt.Println(">> After delete:", find)
	// Output
	// >> After save: Sol1User{ID:1, CreatedAt:2020-10-11 20:40:34.855 +0900 KST, UpdatedAt:2020-10-11 20:40:34.855 +0900 KST, DeletedAtUnix:0, Name:user1}
	// >> After delete: Sol1User{ID:1, CreatedAt:2020-10-11 20:40:34.855 +0900 KST, UpdatedAt:2020-10-11 20:40:34.862 +0900 KST, DeletedAtUnix:1602416434, Name:user1}
}

func NewDB(t *testing.T) *gorm.DB {
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second, // Slow SQL threshold
				LogLevel:      logger.Info, // Log level
				Colorful:      true,
			},
		),
	})
	if err != nil || db == nil {
		t.Fatalf("open db:%v", err)
	}
	models := []interface{}{
		&EmbeddedUser{},
		&Sol1User{},
		&Sol2User{},
		&Sol3User{},
	}

	if err := db.Migrator().DropTable(models...); err != nil {
		t.Fatalf("drop tables:%v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate tables:%v", err)
	}
	return db
}
