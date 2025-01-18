package sm

import "time"

const (
	noState          = ""
	menuState        = "menu"
	editState        = "edit"
	editAddState     = "add"
	editAdd1State    = "add1"
	editAdd2State    = "add2"
	editAddCState    = "addC"
	editDelState     = "del"
	editDelCState    = "delC"
	editChangeState  = "change"
	editChange1State = "change1"
	editChange2State = "change2"
	editChangeCState = "changeC"
	statState        = "stat"
	statShowState    = "show"
	checkState       = "check"
	checkCState      = "checkC"
	settingState     = "setting"
	setting1State    = "setting1"
	settingCState    = "settingC"
	timerState       = "timer"
)

const (
	editButton    = "edit"
	addButton     = "add"
	delButton     = "del"
	changeButton  = "change"
	statButton    = "stat"
	checkButton   = "check"
	settingButton = "setting"
	backButton    = "back"
	prevButton    = "prev"
	nextButton    = "next"
	num1Button    = "num1"
	num2Button    = "num2"
	num3Button    = "num3"
	num4Button    = "num4"
	num5Button    = "num5"
	okButton      = "ok"
	yesButton     = "yes"
	noButton      = "no"
)

var numButtons = []string{num1Button, num2Button, num3Button, num4Button, num5Button}

func buttonNum(button string) int {
	for i, b := range numButtons {
		if b == button {
			return i
		}
	}
	return -1
}

const (
	timeoutTimer int64 = iota
	checkoutTimer
)

const (
	timeoutDuration  = 30 * time.Second
	checkoutDuration = 22 * time.Hour
	dayBound         = 28 * time.Hour
)
