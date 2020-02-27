package game

// these are the different states of the game
const (
	Looking = iota
	Available
	Ready
	Starting
	Countdown
	Racing
	Over
)

// mqtt topics
const (
	TopicHubAvailable = "tinygorace/hub/available"

	TopicRaceAvailable = "tinygorace/race/available"
	TopicRaceStarting  = "tinygorace/race/starting"
	TopicRaceCountdown = "tinygorace/race/countdown"
	TopicRaceStart     = "tinygorace/race/start"
	TopicRaceOver      = "tinygorace/race/over"

	TopicRacerAvailable = "tinygorace/racer/available"
	TopicRacerJoin      = "tinygorace/racer/join"
	TopicRacerReady     = "tinygorace/racer/ready"
	TopicRacerRacing    = "tinygorace/racer/+/racing"

	TopicTrackAvailable = "tinygorace/track/available"
)
