package telegram

import (
	"github.com/go-telegram/bot/models"
	"github.com/tikhomirovv/easyterms/internal/telegram/i18n"
)

func mainMenuKeyboard(locale string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: i18n.T(locale, "btn_new_doc"), CallbackData: cbNewDoc}},
			{
				{Text: i18n.T(locale, "btn_balance"), CallbackData: cbBalance},
				{Text: i18n.T(locale, "btn_buy"), CallbackData: cbBuy},
			},
			{{Text: i18n.T(locale, "btn_demo"), CallbackData: cbDemo}},
		},
	}
}

func draftKeyboard(locale string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: i18n.T(locale, "btn_ready"), CallbackData: cbReadyIngest}},
			{{Text: i18n.T(locale, "btn_new_doc"), CallbackData: cbNewDoc}},
		},
	}
}

func ingestedKeyboard(locale string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: i18n.T(locale, "btn_plain"), CallbackData: cbAnalyzePlain},
				{Text: i18n.T(locale, "btn_highlights"), CallbackData: cbAnalyzeHigh},
			},
			{{Text: i18n.T(locale, "btn_new_doc"), CallbackData: cbNewDoc}},
		},
	}
}

func buyKeyboard(locale string) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: i18n.T(locale, "btn_pkg_1"), CallbackData: cbBuyPkg1}},
			{{Text: i18n.T(locale, "btn_pkg_3"), CallbackData: cbBuyPkg3}},
			{{Text: i18n.T(locale, "btn_pkg_10"), CallbackData: cbBuyPkg10}},
		},
	}
}
