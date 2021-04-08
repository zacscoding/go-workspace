package fixturesexample

import (
	"database/sql"
	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	gMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

type TestSuite struct {
	suite.Suite
	repo *Repository
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	// Open connection
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8&parseTime=True"
	db, err := sql.Open("mysql", dsn)
	s.NoError(err)

	// Setup gorm
	gdb, err := gorm.Open(gMysql.New(gMysql.Config{
		Conn: db,
	}), &gorm.Config{})
	s.NoError(err)
	s.NoError(gdb.AutoMigrate(new(Article)))

	// Setup fixtures
	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.DangerousSkipTestDatabaseCheck(), // disable to check contains "test" in database
		testfixtures.Dialect("mysql"),
		testfixtures.Directory("fixtures"),
	)
	s.NoError(err)
	s.NoError(fixtures.Load())

	// Setup repository
	s.repo = &Repository{db: gdb}
}

func (s *TestSuite) TestFixtures() {
	cases := []struct {
		TCName      string
		Author      string
		ExpectedIds []uint
	}{
		{
			TCName:      "Having multile records",
			Author:      "Griffith Raymond",
			ExpectedIds: []uint{3, 1},
		},
	}
	for _, tc := range cases {
		s.T().Run(tc.TCName, func(t *testing.T) {
			// when
			articles, err := s.repo.FindArticlesByAuthor(tc.Author)
			// then
			assert.NoError(t, err)
			assert.Equal(t, len(tc.ExpectedIds), len(articles))
			for i, article := range articles {
				assert.EqualValues(t, tc.ExpectedIds[i], article.ID)
			}
		})
	}
}
