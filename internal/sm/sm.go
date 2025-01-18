package sm

import (
	"github.com/baldisbk/tgbot/pkg/statemachine"
)

func (d *Doer) TransitionStart(from string) statemachine.Transition {
	return statemachine.Transition{
		Source:      from,
		Destination: menuState,
		Predicate:   CompositePredicate(IsMessage, MessageContains("/start")),
		Callback:    statemachine.CompositeCallback(d.Reset, d.MainMenu),
	}
}

func (d *Doer) TransitionTimeout(from string) statemachine.Transition {
	return statemachine.Transition{
		Source:      from,
		Destination: menuState,
		Predicate:   IsTimer(timeoutTimer),
		Callback:    statemachine.CompositeCallback(d.Reset, d.MainMenu),
	}
}

func SkipMessage(state string) statemachine.Transition {
	return statemachine.Transition{
		Source:      state,
		Destination: state,
		Predicate:   IsMessage,
	}
}

func SkipCallback(state string) statemachine.Transition {
	return statemachine.Transition{
		Source:      state,
		Destination: state,
		Predicate:   IsCallback("*"),
	}
}

func SkipTimeout(state string) statemachine.Transition {
	return statemachine.Transition{
		Source:      state,
		Destination: state,
		Predicate:   IsTimer(timeoutTimer),
	}
}

func MakeStateMachine(doer *Doer) statemachine.Machine {
	state := doer.User.State
	if state == "" {
		state = noState
	}
	return statemachine.NewSM(state, []statemachine.Transition{
		// start
		{
			Source:      noState,
			Destination: menuState,
			Predicate:   CompositePredicate(IsMessage, MessageContains("/start")),
			Callback:    statemachine.CompositeCallback(doer.Greeting, doer.SetCheckout, doer.MainMenu),
		},
		// main menu
		{
			Source:      menuState,
			Destination: editState,
			Predicate:   IsCallback(editButton),
			Callback:    doer.EditMenu,
		},
		{
			Source:      menuState,
			Destination: checkState,
			Predicate:   IsCallback(checkButton),
			Callback:    doer.CheckMenu,
		},
		{
			Source:      menuState,
			Destination: statState,
			Predicate:   IsCallback(statButton),
			Callback:    doer.StatMenu,
		},
		{
			Source:      menuState,
			Destination: settingState,
			Predicate:   IsCallback(settingButton),
			Callback:    doer.MainMenu, // doer.SettingsMenu,
		},
		doer.TransitionStart(menuState),
		SkipMessage(menuState),
		SkipCallback(menuState),
		SkipTimeout(menuState),
		// edit menu
		{
			Source:      editState,
			Destination: editChangeState,
			Predicate:   IsCallback(changeButton),
			Callback:    doer.T(doer.ChangeMenu),
		},
		{
			Source:      editState,
			Destination: editDelState,
			Predicate:   IsCallback(delButton),
			Callback:    doer.T(doer.DeleteMenu),
		},
		{
			Source:      editState,
			Destination: editAddState,
			Predicate:   IsCallback(addButton),
			Callback:    doer.T(doer.Add),
		},
		{
			Source:      editState,
			Destination: menuState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.MainMenu,
		},
		doer.TransitionStart(editState),
		doer.TransitionTimeout(editState),
		SkipMessage(editState),
		SkipCallback(editState),
		// add scenario - input name
		{
			Source:      editAddState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.T(doer.EditMenu),
		},
		{
			Source:      editAddState,
			Destination: editAddCState,
			Predicate:   IsMessage,
			Callback:    doer.T(doer.Msg(doer.AddConfirm)),
		},
		doer.TransitionTimeout(editAddState),
		SkipCallback(editAddState),
		// add scenario - confirm
		{
			Source:      editAddCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.DoAdd,
				doer.EditMenu,
			)),
		},
		{
			Source:      editAddCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.T(doer.EditMenu),
		},
		doer.TransitionStart(editAddCState),
		doer.TransitionTimeout(editAddCState),
		SkipMessage(editAddCState),
		SkipCallback(editAddCState),
		// change scenario - select
		{
			Source:      editChangeState,
			Destination: editChangeState,
			Predicate:   IsCallback(nextButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.ChangeMenu,
			)),
		},
		{
			Source:      editChangeState,
			Destination: editChangeState,
			Predicate:   IsCallback(prevButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.ChangeMenu,
			)),
		},
		{
			Source:      editChangeState,
			Destination: editChange1State,
			Predicate:   IsNumButton,
			Callback:    doer.T(doer.Cb(doer.Change)),
		},
		{
			Source:      editChangeState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.T(doer.EditMenu),
		},
		doer.TransitionStart(editChangeState),
		doer.TransitionTimeout(editChangeState),
		SkipMessage(editChangeState),
		SkipCallback(editChangeState),
		// change scenario - input name
		{
			Source:      editChange1State,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.T(doer.EditMenu),
		},
		{
			Source:      editChange1State,
			Destination: editChangeCState,
			Predicate:   IsMessage,
			Callback:    doer.T(doer.Msg(doer.ChangeConfirm)),
		},
		doer.TransitionTimeout(editChangeState),
		SkipCallback(editChangeState),
		// change scenario - confirm
		{
			Source:      editChangeCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.DoAdd,
				doer.EditMenu,
			)),
		},
		{
			Source:      editChangeCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.T(doer.EditMenu),
		},
		doer.TransitionStart(editChangeCState),
		doer.TransitionTimeout(editChangeCState),
		SkipMessage(editChangeCState),
		SkipCallback(editChangeCState),
		// delete scenario - select
		{
			Source:      editDelState,
			Destination: editDelState,
			Predicate:   IsCallback(nextButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.DeleteMenu,
			)),
		},
		{
			Source:      editDelState,
			Destination: editDelState,
			Predicate:   IsCallback(prevButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.DeleteMenu,
			)),
		},
		{
			Source:      editDelState,
			Destination: editDelCState,
			Predicate:   IsNumButton,
			Callback:    doer.T(doer.Cb(doer.DelConfirm)),
		},
		{
			Source:      editDelState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.T(doer.EditMenu),
		},
		doer.TransitionStart(editDelState),
		doer.TransitionTimeout(editDelState),
		SkipMessage(editDelState),
		SkipCallback(editDelState),
		// delete scenario - confirm
		{
			Source:      editDelCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.DoDel,
				doer.EditMenu,
			)),
		},
		{
			Source:      editDelCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.T(doer.EditMenu),
		},
		doer.TransitionStart(editDelCState),
		doer.TransitionTimeout(editDelCState),
		SkipMessage(editDelCState),
		SkipCallback(editDelCState),
		// check scenario - select
		{
			Source:      checkState,
			Destination: checkState,
			Predicate:   IsCallback(nextButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.CheckMenu,
			)),
		},
		{
			Source:      checkState,
			Destination: checkState,
			Predicate:   IsCallback(prevButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.CheckMenu,
			)),
		},
		{
			Source:      checkState,
			Destination: checkCState,
			Predicate:   IsNumButton,
			Callback:    doer.T(doer.Cb(doer.CheckConfirm)),
		},
		{
			Source:      checkState,
			Destination: menuState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.MainMenu,
		},
		doer.TransitionStart(checkState),
		doer.TransitionTimeout(checkState),
		SkipMessage(checkState),
		SkipCallback(checkState),
		// check scenario - confirm
		{
			Source:      checkCState,
			Destination: menuState,
			Predicate:   IsCallback(yesButton),
			Callback: statemachine.CompositeCallback(
				doer.DoCheck,
				doer.MainMenu,
			),
		},
		{
			Source:      checkCState,
			Destination: menuState,
			Predicate:   IsCallback(noButton),
			Callback: statemachine.CompositeCallback(
				doer.DoNoCheck,
				doer.MainMenu,
			),
		},
		{
			Source:      checkState,
			Destination: menuState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.MainMenu,
		},
		doer.TransitionStart(checkCState),
		doer.TransitionTimeout(checkCState),
		SkipMessage(checkCState),
		SkipCallback(checkCState),
		// stat scenario - select
		{
			Source:      statState,
			Destination: statState,
			Predicate:   IsCallback(nextButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.StatMenu,
			)),
		},
		{
			Source:      statState,
			Destination: statState,
			Predicate:   IsCallback(prevButton),
			Callback: doer.T(statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.StatMenu,
			)),
		},
		{
			Source:      statState,
			Destination: statShowState,
			Predicate:   IsNumButton,
			Callback:    doer.T(doer.Cb(doer.ShowStat)),
		},
		{
			Source:      statState,
			Destination: menuState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.MainMenu,
		},
		doer.TransitionStart(statState),
		doer.TransitionTimeout(statState),
		SkipMessage(statState),
		SkipCallback(statState),
		// stat scenario - show
		{
			Source:      statShowState,
			Destination: statState,
			Predicate:   IsCallback(okButton),
			Callback:    doer.T(doer.StatMenu),
		},
		doer.TransitionStart(statShowState),
		doer.TransitionTimeout(statShowState),
		SkipMessage(statShowState),
		SkipCallback(statShowState),
	}, true)
}
