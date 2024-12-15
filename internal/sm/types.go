package sm

import (
	"encoding/json"
	"strconv"

	"github.com/baldisbk/tgeasybot/internal/orm"
	"golang.org/x/xerrors"
)

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
}

type Message struct {
	From User   `json:"from"`
	Text string `json:"text"`
	Date int64  `json:"date"`
}

type CallbackQuery struct {
	From   User   `json:"from"`
	Id     string `json:"id"`
	Button string `json:"button"`
}

type Timer struct {
	ID int64
}

func UnmarshalEvent(event *orm.Event) (interface{}, error) {
	switch event.Type {
	case orm.EventTypeMessage:
		var msg *Message
		if err := json.Unmarshal([]byte(event.Payload), &msg); err != nil {
			return nil, xerrors.Errorf("unmarshal messsage: %w", err)
		}
		return msg, nil
	case orm.EventTypeCallback:
		var cb *CallbackQuery
		if err := json.Unmarshal([]byte(event.Payload), &cb); err != nil {
			return nil, xerrors.Errorf("unmarshal callback: %w", err)
		}
		return cb, nil
	case orm.EventTypeTimer:
		id, err := strconv.ParseInt(event.Payload, 0, 64)
		if err != nil {
			return nil, xerrors.Errorf("unmarshal timer: %w", err)
		}
		return &Timer{ID: id}, nil
	}
	return nil, xerrors.Errorf("unknown type: %s", event.Type)
}
