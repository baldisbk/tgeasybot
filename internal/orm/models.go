package orm

const (
	EventTypeMessage  = "message"
	EventTypeCallback = "callback"
	EventTypeTimer    = "timer"
)

type User struct {
	ID       int64
	Name     string
	Busy     int64 // timestamp
	State    string
	Contents string
}

type Event struct {
	ID      int64 // sequental order for this user; negative for timers to non-intersect
	TS      int64 // rough timestamp (seconds)
	UserID  int64
	Busy    int64 // timestamp
	Type    string
	Payload string

	UserName string // not in DB, but sent by TGAPI
}

type State struct {
	Offset int64
}

type Timer struct {
	ID         int64 // unique timer ID for this user; goes to event ID
	TS         int64 // rough timestamp (seconds)
	UserID     int64
	Repeatable bool
}
