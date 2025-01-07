package orm

const (
	eventTable = `
CREATE TABLE IF NOT EXISTS events (
	id BIGINT,
	ts BIGINT,
	userId BIGINT,
	busy BIGINT,
	type TEXT,
	payload TEXT,
	PRIMARY KEY (userId, ts, id)
);
CREATE INDEX IF NOT EXISTS event_index ON events USING btree (userId, ts, id);
`

	userTable = `
CREATE TABLE IF NOT EXISTS users (
	id BIGINT PRIMARY KEY,
	name TEXT,
	busy BIGINT,
	state TEXT,
	contents TEXT
);`

	timerTable = `
CREATE TABLE IF NOT EXISTS timers (
	id BIGINT,
	ts BIGINT,
	userId BIGINT,
	repeatable BOOL,
	PRIMARY KEY (userId, id)
);
CREATE INDEX IF NOT EXISTS timer_index ON timers USING btree (ts, userId, id);
`

	stateTable = `
CREATE TABLE IF NOT EXISTS state (
	id INT PRIMARY KEY,
	lastOffset BIGINT,
	config TEXT
);`

	// fetch updates
	selectState = `SELECT lastOffset, config FROM state;`
	insertEvent = `
INSERT INTO events (id, ts, userId, busy, type, payload)
VALUES ($1, $2, $3, 0, $4, $5)
ON CONFLICT DO NOTHING;`
	insertUser = `
INSERT INTO users (id, name, busy, state, contents)
VALUES ($1, $2, 0, $3, $4)
ON CONFLICT (id) DO UPDATE SET name=$2;`
	updateState = `INSERT INTO state (id, lastOffset, config) VALUES (0, $1, $2) ON CONFLICT (id) DO UPDATE SET lastOffset=$1, config=$2;`

	// fetch events
	listEvents = `SELECT id, ts, userId, type, payload FROM events WHERE busy < $1;`
	selectUser = `SELECT name, busy, state, contents FROM users WHERE id=$1;`
	markUser   = `UPDATE users SET busy = $2 WHERE id = $1;`
	markEvent  = `UPDATE events SET busy = $4 WHERE id = $1 AND ts = $2 AND userId = $3;`

	// process events
	updateUser  = `UPDATE users SET name = $2, state = $3, busy = 0, contents = $4 WHERE id = $1;`
	deleteEvent = `DELETE FROM events WHERE id = $1 AND ts = $2 AND userId = $3;`

	// set timer
	setTimer = `INSERT INTO timers (id, ts, userId, repeatable) VALUES ($1, $2, $3, $4) ON CONFLICT (userId, id) DO UPDATE SET ts=$2, repeatable=$4;`

	// process timers
	fetchTimers = `SELECT id, ts, userId, repeatable FROM timers WHERE ts < $1;`
	deleteTimer = `DELETE FROM timers WHERE id = $1 AND userId = $2;`
)
