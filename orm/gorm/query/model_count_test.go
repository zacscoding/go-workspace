package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountQuery(t *testing.T) {
	u11 := &Account{
		Name: "user1",
	}
	u12 := &Account{
		Name: "user1",
	}
	u21 := &Account{
		Name: "user2",
	}
	assert.NoError(t, db.Save(u11).Error)
	assert.NoError(t, db.Save(u12).Error)
	assert.NoError(t, db.Save(u21).Error)

	// count
	var count int
	count = 0
	assert.NoError(t, db.Model(&Account{}).Where("name = ?", "user1").Count(&count).Error)
	assert.Equal(t, 2, count)

	count = 0
	assert.NoError(t, db.Model(&Account{}).Where("name = ?", "user2").Count(&count).Error)
	assert.Equal(t, 1, count)

	count = 0
	assert.NoError(t, db.Model(&Account{}).Where("name = ?", "user3").Count(&count).Error)
	assert.Equal(t, 0, count)
}
