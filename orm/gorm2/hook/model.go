package hook

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type HookUser struct {
	ID        uint      `gorm:"column:id;primarykey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	beforeSleep time.Duration `gorm:"-"`
	afterSleep  time.Duration `gorm:"-"`
	afterError  error         `gorm:"-"`
}

func (u *HookUser) BeforeCreate(tx *gorm.DB) error {
	log.Printf("[Goroutine-%d] %s BeforeCreate is called", goroutineID(), u.Name)
	if u.beforeSleep != 0 {
		time.Sleep(u.beforeSleep)
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = u.CreatedAt
	return nil
}

func (u *HookUser) AfterCreate(tx *gorm.DB) error {
	log.Printf("[Goroutine-%d] %s After is called", goroutineID(), u.Name)
	if u.afterSleep != 0 {
		time.Sleep(u.afterSleep)
	}
	if u.afterError != nil {
		return u.afterError
	}
	return nil
}

func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
