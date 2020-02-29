package game

// these are the different states of the game
const (
	Looking = iota
	Available
	Ready
	Starting
	Countdown
	Start
	Over
	Winner
)

// mqtt topics
const (
	TopicHubAvailable = "tinygorace/hub/available"

	TopicRaceAvailable = "tinygorace/race/available"
	TopicRaceStarting  = "tinygorace/race/starting"
	TopicRaceJoined    = "tinygorace/race/racer/+/joined"
	TopicRaceCountdown = "tinygorace/race/countdown"
	TopicRaceStart     = "tinygorace/race/start"
	TopicRaceOver      = "tinygorace/race/over"
	TopicRaceWinner    = "tinygorace/race/winner"

	TopicRacerAvailable = "tinygorace/racer/+/available"
	TopicRacerJoin      = "tinygorace/racer/+/join"
	TopicRacerReady     = "tinygorace/racer/+/ready"
	TopicRacerRacing    = "tinygorace/racer/+/racing"
	TopicRacerPosition  = "tinygorace/racer/+/position"

	TopicTrackAvailable = "tinygorace/track/available"
)

const (
	// Accelleration how fast the racer moves per tap
	Accelleration = 2.0

	// Friction how much the racer slows down per 100 ms
	Friction = 0.15
)

// Racer is one of the racers on the track.
type Racer struct {
	// Speed is how fast the racer is going
	Speed int

	// Pos is the position of the racer on the track
	Pos int

	// Laps is how many laps the racer has completed
	Laps int
}

// Track is the track.
type Track struct {
	// Len is how long the track is
	Len int
}
