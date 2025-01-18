package sm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	now, err := time.Parse(time.DateTime, "")
	require.NoError(t, err)
	testCases := []struct {
		name string
		ach  UserStat
		set  UserSettings
		exp  string
	}{
		{
			name: "",
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.exp, achievementFormat(c.ach, now, c.set))
		})
	}
}
