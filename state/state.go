package state

type State int

const (
	Interference State = 1 + iota
	Power
	Battery
	Unknown
)

var states = [...]string{
	"RF_INT_DET",
	"TX_TYPE",
	"BATT",
	"UNKNOWN",
}

func (s State) String() string {
	return states[s-1]
}
