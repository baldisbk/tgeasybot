package sm

import (
	"strconv"

	"github.com/baldisbk/tgbot/pkg/tgapi"
)

var (
	mainMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{
			{Text: "Check", CallbackData: checkButton},
			{Text: "Edit", CallbackData: editButton},
		},
		{
			{Text: "Statistics", CallbackData: statButton},
			{Text: "Settings", CallbackData: settingButton},
		},
	}}
	editMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{
			{Text: "Add", CallbackData: addButton},
			{Text: "Delete", CallbackData: delButton},
			{Text: "Change", CallbackData: changeButton},
		},
		{
			{Text: "Back", CallbackData: backButton},
		},
	}}
	okMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{{Text: "OK", CallbackData: okButton}},
	}}
	backMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{{Text: "Back", CallbackData: backButton}},
	}}
	yesNoMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{
			{Text: "Yes", CallbackData: yesButton},
			{Text: "No", CallbackData: noButton},
		},
	}}
	yesNoBackMenu = tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		{
			{Text: "Yes", CallbackData: yesButton},
			{Text: "No", CallbackData: noButton},
		},
		{{Text: "Back", CallbackData: backButton}},
	}}
)

func listMenu(from, size int) tgapi.InlineKeyboard {
	line := []tgapi.InlineKeyboardButton{}
	if from > 0 {
		line = append(line, tgapi.InlineKeyboardButton{Text: "<", CallbackData: prevButton})
	}
	for i := from; i < size && i-from < len(numButtons); i++ {
		line = append(line, tgapi.InlineKeyboardButton{
			Text:         strconv.FormatInt(int64(i+1), 10),
			CallbackData: numButtons[i-from],
		})
	}
	if size-from > len(numButtons) {
		line = append(line, tgapi.InlineKeyboardButton{Text: ">", CallbackData: nextButton})
	}
	return tgapi.InlineKeyboard{InlineKeyboard: [][]tgapi.InlineKeyboardButton{
		line,
		{{Text: "Back", CallbackData: backButton}},
	}}
}
