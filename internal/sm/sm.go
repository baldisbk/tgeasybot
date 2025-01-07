package sm

import (
	"github.com/baldisbk/tgbot/pkg/statemachine"
)

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
			Callback: statemachine.CompositeCallback(
				doer.Msg(doer.Greeting),
				doer.MainMenu,
			),
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
			Callback:    doer.MainMenu, // doer.CheckMenu,
		},
		{
			Source:      menuState,
			Destination: statState,
			Predicate:   IsCallback(statButton),
			Callback:    doer.MainMenu, // doer.StatMenu,
		},
		{
			Source:      menuState,
			Destination: settingState,
			Predicate:   IsCallback(settingButton),
			Callback:    doer.MainMenu, // doer.SettingsMenu,
		},
		// edit menu
		{
			Source:      editState,
			Destination: editChangeState,
			Predicate:   IsCallback(changeButton),
			Callback:    doer.ChangeMenu,
		},
		{
			Source:      editState,
			Destination: editDelState,
			Predicate:   IsCallback(delButton),
			Callback:    doer.DeleteMenu,
		},
		{
			Source:      editState,
			Destination: editAddState,
			Predicate:   IsCallback(addButton),
			Callback:    doer.Add,
		},
		{
			Source:      editState,
			Destination: menuState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.MainMenu,
		},
		// add scenario
		{
			Source:      editAddState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.EditMenu,
		},
		{
			Source:      editAddState,
			Destination: editAddCState,
			Predicate:   IsMessage,
			Callback:    doer.Msg(doer.AddConfirm),
		},
		{
			Source:      editAddCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: statemachine.CompositeCallback(
				doer.DoAdd,
				doer.EditMenu,
			),
		},
		{
			Source:      editAddCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.EditMenu,
		},
		// change scenario
		{
			Source:      editChangeState,
			Destination: editChangeState,
			Predicate:   IsCallback(nextButton),
			Callback: statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.ChangeMenu,
			),
		},
		{
			Source:      editChangeState,
			Destination: editChangeState,
			Predicate:   IsCallback(prevButton),
			Callback: statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.ChangeMenu,
			),
		},
		{
			Source:      editChangeState,
			Destination: editChange1State,
			Predicate:   IsNumButton,
			Callback:    doer.Cb(doer.Change),
		},
		{
			Source:      editChangeState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.EditMenu,
		},
		{
			Source:      editChange1State,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.EditMenu,
		},
		{
			Source:      editChange1State,
			Destination: editChangeCState,
			Predicate:   IsMessage,
			Callback:    doer.Msg(doer.ChangeConfirm),
		},
		{
			Source:      editChangeCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: statemachine.CompositeCallback(
				doer.DoAdd,
				doer.EditMenu,
			),
		},
		{
			Source:      editChangeCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.EditMenu,
		},
		// delete scenario
		{
			Source:      editDelState,
			Destination: editDelState,
			Predicate:   IsCallback(nextButton),
			Callback: statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.DeleteMenu,
			),
		},
		{
			Source:      editDelState,
			Destination: editDelState,
			Predicate:   IsCallback(prevButton),
			Callback: statemachine.CompositeCallback(
				doer.Cb(doer.Scroll),
				doer.DeleteMenu,
			),
		},
		{
			Source:      editDelState,
			Destination: editDelCState,
			Predicate:   IsNumButton,
			Callback:    doer.Cb(doer.DelConfirm),
		},
		{
			Source:      editDelState,
			Destination: editState,
			Predicate:   IsCallback(backButton),
			Callback:    doer.EditMenu,
		},
		{
			Source:      editDelCState,
			Destination: editState,
			Predicate:   IsCallback(yesButton),
			Callback: statemachine.CompositeCallback(
				doer.DoDel,
				doer.EditMenu,
			),
		},
		{
			Source:      editDelCState,
			Destination: editState,
			Predicate:   IsCallback(noButton),
			Callback:    doer.EditMenu,
		},
	}, true)
}
