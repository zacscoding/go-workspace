package decodeexample

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	synchronizedResult = `false`
	syncingResult      = `{"startOffset": 10, "currentOffset": 25, "endOffset": 30}`
)

type SyncingStatus struct {
	Syncing       bool `json:"-"`
	StartOffset   int  `json:"startOffset"`
	CurrentOffset int  `json:"currentOffset"`
	EndOffset     int  `json:"endOffset"`
}

func (s *SyncingStatus) UnmarshalJSON(data []byte) error {
	if "false" == string(data) {
		s.Syncing = false
		return nil
	}
	type syncStatus struct {
		StartOffset   int `json:"startOffset"`
		CurrentOffset int `json:"currentOffset"`
		EndOffset     int `json:"endOffset"`
	}
	var status syncStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}
	s.Syncing = true
	s.StartOffset = status.StartOffset
	s.CurrentOffset = status.CurrentOffset
	s.EndOffset = status.EndOffset
	return nil
}

func TestSyncingStatusUnmarshal(t *testing.T) {
	t.Run("Not Syncing", func(t *testing.T) {
		var status SyncingStatus
		err := json.Unmarshal([]byte(synchronizedResult), &status)
		assert.NoError(t, err)
		assert.False(t, status.Syncing)
		assert.EqualValues(t, 0, status.StartOffset)
		assert.EqualValues(t, 0, status.CurrentOffset)
		assert.EqualValues(t, 0, status.EndOffset)
	})

	t.Run("Syncing", func(t *testing.T) {
		var status SyncingStatus
		err := json.Unmarshal([]byte(syncingResult), &status)
		assert.NoError(t, err)
		assert.True(t, status.Syncing)
		assert.EqualValues(t, 10, status.StartOffset)
		assert.EqualValues(t, 25, status.CurrentOffset)
		assert.EqualValues(t, 30, status.EndOffset)
	})
}
