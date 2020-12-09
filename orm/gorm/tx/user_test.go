package tx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/stretchr/testify/suite"
	"strings"
	"sync"
	"testing"
	"time"
)

type UserTxTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *UserTxTestSuite) SetupSuite() {
	db, err := gorm.Open("mysql", "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True")
	s.NoError(err)

	s.NoError(db.DropTableIfExists(new(AccountModel)).Error)
	s.NoError(db.CreateTable(new(AccountModel)).Error)

	s.db = db
}

func (s *UserTxTestSuite) SetupTest() {
	s.NoError(s.db.Unscoped().Delete(new(AccountModel)).Error)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(UserTxTestSuite))
}

var userPrefixes = []string{"Alice", "Bob", "Zac", "John", "Brown", "Gail", "Messi"}

func (s *UserTxTestSuite) TestTemp() {
	readCtx, cancel := context.WithCancel(context.Background())
	go loopReadAccounts(readCtx, s.db, time.Second*2)

	wg := sync.WaitGroup{}
	writer := 3
	for i := 0; i < writer; i++ {
		wg.Add(1)
		userPrefix := userPrefixes[i]
		go func() {
			defer wg.Done()
			writeAccounts(s.db, 10, time.Second*5, userPrefix)
		}()
	}
	wg.Wait()
	time.Sleep(time.Second * 2)
	cancel()
}

func writeAccounts(db *gorm.DB, repeat int, duration time.Duration, prefix string) {
	InTxWithContext(context.Background(), db, &sql.TxOptions{
		Isolation: sql.LevelReadUncommitted,
		ReadOnly:  false,
	}, func(txDB *gorm.DB) error {
		for i := 0; i < repeat; i++ {
			details, err := readAccounts(txDB)
			if err != nil {
				fmt.Printf("## [Writer-%s] Failed to read accounts:%v\n", prefix, err)
			} else {
				fmt.Printf("## [Writer-%s] Success to read accounts:%s\n", prefix, strings.Join(details, ","))
			}
			if err := txDB.Create(&AccountModel{
				Email:    fmt.Sprintf("%s-%d@email.com", prefix, i),
				Username: fmt.Sprintf("%s-%d", prefix, i),
				Age:      i,
			}).Error; err != nil {
				fmt.Printf("## [Writer-%s] Failed to save a new account:%v\n", prefix, err)
				return err
			}
			fmt.Printf("## [Writer-%s] Success to save a new account-%d\n", prefix, i)
			time.Sleep(duration)
		}
		return nil
	})
}

func loopReadAccounts(ctx context.Context, db *gorm.DB, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	readFunc := func() {
		details, err := readAccounts(db)
		if err != nil {
			fmt.Printf("[Reader] Failed to read accounts. err:%v\n", err)
			return
		}
		if len(details) == 0 {
			fmt.Println("[Reader] Read accounts: empty")
			return
		}
		fmt.Printf("#[Reader] Read accounts: %s\n", strings.Join(details, ","))
	}

	for {
		select {
		case <-ticker.C:
			readFunc()
		case <-ctx.Done():
			readFunc() // last read
			return
		}
	}
}

func readAccounts(db *gorm.DB) ([]string, error) {
	var accounts []*AccountModel
	err := db.Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	var details []string
	for _, account := range accounts {
		details = append(details, fmt.Sprintf("[%s,%s]", account.Email, account.Username))
	}
	if len(details) == 0 {
		return []string{}, nil
	}
	return details, nil
}
