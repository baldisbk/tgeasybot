package sm

import (
	"encoding/json"
	"time"
)

type Record struct {
	Date   time.Time `json:"date,omitempty"`
	Result bool      `json:"result,omitempty"`
}

type UserStat struct {
	Name     string   `json:"name,omitempty"`
	Question string   `json:"question,omitempty"`
	History  []Record `json:"history,omitempty"`
}

type UserPermanentState struct {
	Stats []UserStat `json:"stats,omitempty"`
}

type UserTemporaryState struct {
	StatIndex   int      `json:"stat_index,omitempty"`
	InputStat   UserStat `json:"input_stat,omitempty"`
	LastMessage uint64   `json:"last_message,omitempty"`
}

type UserSettings struct {
	CheckupOffset time.Duration `json:"checkup_offset,omitempty"`
	DayBound      time.Duration `json:"day_bound,omitempty"`
}

type UserState struct {
	UserPermanentState
	UserTemporaryState
	UserSettings
}

func (s *UserState) Marshal() (string, error) {
	b, err := json.Marshal(*s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func StateUnmarshal(from string) (*UserState, error) {
	s := UserState{
		UserSettings: UserSettings{
			CheckupOffset: checkoutDuration,
			DayBound:      dayBound,
		},
	}
	if from == "" {
		return &s, nil
	}
	err := json.Unmarshal([]byte(from), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
