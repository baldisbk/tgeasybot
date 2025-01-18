package sm

import (
	"fmt"
	"time"
)

func achievementFormat(ach UserStat, from time.Time, settings UserSettings) string {
	var first, firstRow time.Time
	var week, month, row, lastDays int
	now := from.Truncate(24 * time.Hour)
	for i, r := range ach.History {
		if !r.Result {
			continue
		}
		if i == 0 {
			first = r.Date
			firstRow = r.Date
		}
		days := int(r.Date.Sub(now) / (24 * time.Hour))
		if lastDays != 0 && days-lastDays > 1 {
			row = 0
			firstRow = r.Date
		}
		if days <= 30 {
			month++
		}
		if days <= 7 {
			week++
		}
		row++
	}
	return fmt.Sprintf("Achievement %q:\nStarted at %s\n%d for last week\n%d for last month\n%d in a row since %s.",
		ach.Name, first.Format(time.DateOnly), week, month, row, firstRow.Format(time.DateOnly))
}
