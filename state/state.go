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

type Intf int

const (
	None Intf = 1 + iota
	Critical
)

var intfs = [...]string{
	"NONE",
	"CRITICAL",
}

func (i Intf) String() string {
	return intfs[i-1]
}

type BattState int

const (
	Cycles BattState = 1 + iota
	RunTime
	Type
	Charge
	Bars
)

var batt = [...]string{
	"BATT_CYCLE",
	"BATT_RUN_TIME",
	"BATT_TYPE",
	"BATT_CHARGE",
	"BATT_BARS",
}

func (b BattState) String() string {
	return batt[b-1]
}
