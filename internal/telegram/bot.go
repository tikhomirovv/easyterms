package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/tikhomirovv/easyterms/internal/core/services/analysis"
	"github.com/tikhomirovv/easyterms/internal/telegram/i18n"
)

// Run starts the Telegram bot until ctx is cancelled.
func Run(ctx context.Context, token string, app *App) error {
	tb, err := bot.New(token, bot.WithDefaultHandler(app.handleDefault))
	if err != nil {
		return fmt.Errorf("telegram bot: %w", err)
	}

	tb.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, app.wrap(app.handleStart))
	tb.RegisterHandler(bot.HandlerTypeMessageText, "/new", bot.MatchTypeExact, app.wrap(app.handleNewDoc))
	tb.RegisterHandler(bot.HandlerTypeMessageText, "/demo", bot.MatchTypeExact, app.wrap(app.handleDemo))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbNewDoc, bot.MatchTypeExact, app.wrap(app.handleNewDoc))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbReadyIngest, bot.MatchTypeExact, app.wrap(app.handleReady))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAnalyzePlain, bot.MatchTypeExact, app.wrap(app.handleAnalyzePlain))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbAnalyzeHigh, bot.MatchTypeExact, app.wrap(app.handleAnalyzeHighlights))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbBalance, bot.MatchTypeExact, app.wrap(app.handleBalance))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbBuy, bot.MatchTypeExact, app.wrap(app.handleBuy))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbDemo, bot.MatchTypeExact, app.wrap(app.handleDemo))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbBuyPkg1, bot.MatchTypeExact, app.wrap(app.handleBuyPkg1))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbBuyPkg3, bot.MatchTypeExact, app.wrap(app.handleBuyPkg3))
	tb.RegisterHandler(bot.HandlerTypeCallbackQueryData, cbBuyPkg10, bot.MatchTypeExact, app.wrap(app.handleBuyPkg10))

	app.log.Info("telegram bot listening")
	tb.Start(ctx)
	return nil
}

func (a *App) wrap(fn func(context.Context, *bot.Bot, *models.Update) error) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if err := fn(ctx, b, update); err != nil {
			a.log.Error("handler", slog.String("error", err.Error()))
		}
	}
}

func (a *App) handleDefault(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}
	_ = a.handleText(ctx, b, update)
}

func (a *App) handleStart(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	_, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	return a.reply(ctx, b, update, withDisclaimer(locale, i18n.T(locale, "welcome")), mainMenuKeyboard(locale))
}

func (a *App) handleDemo(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	return a.reply(ctx, b, update, withDisclaimer(locale, i18n.T(locale, "demo")), mainMenuKeyboard(locale))
}

func (a *App) handleNewDoc(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	if _, err := a.docs.CreateDocument(ctx, user.ID); err != nil {
		return err
	}
	return a.reply(ctx, b, update, i18n.T(locale, "new_doc_created"), draftKeyboard(locale))
}

func (a *App) handleText(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	draft, err := a.activeDraft(ctx, user.ID)
	if err != nil {
		return a.reply(ctx, b, update, i18n.T(locale, "no_active_doc"), mainMenuKeyboard(locale))
	}

	text := update.Message.Text
	if isURL(text) {
		err = a.docs.AddURLSource(ctx, user.ID, draft.ID, text)
	} else {
		err = a.docs.AddTextSource(ctx, user.ID, draft.ID, text)
	}
	if err != nil {
		return a.reply(ctx, b, update, userFacingErr(locale, err), draftKeyboard(locale))
	}
	return a.reply(ctx, b, update, i18n.T(locale, "content_added"), draftKeyboard(locale))
}

func (a *App) handleReady(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	draft, err := a.activeDraft(ctx, user.ID)
	if err != nil {
		return a.reply(ctx, b, update, i18n.T(locale, "no_active_doc"), mainMenuKeyboard(locale))
	}
	_, err = a.docs.Ingest(ctx, user.ID, draft.ID)
	if err != nil {
		return a.reply(ctx, b, update, userFacingErr(locale, err), draftKeyboard(locale))
	}
	return a.reply(ctx, b, update, withDisclaimer(locale, i18n.T(locale, "ingest_ok")), ingestedKeyboard(locale))
}

func (a *App) handleAnalyzePlain(ctx context.Context, b *bot.Bot, update *models.Update) error {
	return a.runAnalysis(ctx, b, update, analysis.TypePlain, "analysis_plain", formatPlainPayload)
}

func (a *App) handleAnalyzeHighlights(ctx context.Context, b *bot.Bot, update *models.Update) error {
	return a.runAnalysis(ctx, b, update, analysis.TypeHighlights, "analysis_highlights", formatHighlightsPayload)
}

func (a *App) runAnalysis(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
	analysisType, msgKey string,
	format func([]byte) string,
) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	doc, err := a.latestIngested(ctx, user.ID)
	if err != nil {
		return a.reply(ctx, b, update, i18n.T(locale, "no_active_doc"), mainMenuKeyboard(locale))
	}
	res, err := a.analysis.Run(ctx, user.ID, doc.ID, analysisType)
	if err != nil {
		return a.reply(ctx, b, update, userFacingErr(locale, err), ingestedKeyboard(locale))
	}
	body := format(res.Payload)
	if len(body) > 3500 {
		body = body[:3500] + "…"
	}
	text := fmt.Sprintf(i18n.T(locale, msgKey), body)
	return a.reply(ctx, b, update, withDisclaimer(locale, text), ingestedKeyboard(locale))
}

func (a *App) handleBalance(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	bal, err := a.billing.Balance(ctx, user.ID)
	if err != nil {
		return err
	}
	return a.reply(ctx, b, update, fmt.Sprintf(i18n.T(locale, "balance"), bal), mainMenuKeyboard(locale))
}

func (a *App) handleBuy(ctx context.Context, b *bot.Bot, update *models.Update) error {
	locale := a.locale(update)
	return a.reply(ctx, b, update, i18n.T(locale, "buy_intro"), buyKeyboard(locale))
}

func (a *App) handleBuyPkg1(ctx context.Context, b *bot.Bot, update *models.Update) error {
	return a.handleBuyPackage(ctx, b, update, "checks_1")
}
func (a *App) handleBuyPkg3(ctx context.Context, b *bot.Bot, update *models.Update) error {
	return a.handleBuyPackage(ctx, b, update, "checks_3")
}
func (a *App) handleBuyPkg10(ctx context.Context, b *bot.Bot, update *models.Update) error {
	return a.handleBuyPackage(ctx, b, update, "checks_10")
}

func (a *App) handleBuyPackage(ctx context.Context, b *bot.Bot, update *models.Update, packageID string) error {
	locale := a.locale(update)
	user, err := a.ensureUser(ctx, a.telegramID(update), locale)
	if err != nil {
		return err
	}
	msg, err := a.startPurchase(ctx, user.ID, packageID, locale)
	if err != nil {
		return a.reply(ctx, b, update, userFacingErr(locale, err), mainMenuKeyboard(locale))
	}
	return a.reply(ctx, b, update, msg, mainMenuKeyboard(locale))
}

func (a *App) reply(ctx context.Context, b *bot.Bot, update *models.Update, text string, kb *models.InlineKeyboardMarkup) error {
	chatID := a.chatID(update)
	if chatID == 0 {
		return fmt.Errorf("missing chat id")
	}
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	}
	if kb != nil {
		params.ReplyMarkup = kb
	}
	_, err := b.SendMessage(ctx, params)
	if update.CallbackQuery != nil {
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
		})
	}
	return err
}

func (a *App) locale(update *models.Update) string {
	if update.Message != nil && update.Message.From != nil && update.Message.From.LanguageCode != "" {
		return i18n.LocaleFromTelegram(update.Message.From.LanguageCode)
	}
	if update.CallbackQuery != nil && update.CallbackQuery.From.LanguageCode != "" {
		return i18n.LocaleFromTelegram(update.CallbackQuery.From.LanguageCode)
	}
	return "en"
}

func (a *App) telegramID(update *models.Update) int64 {
	if update.Message != nil && update.Message.From != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func (a *App) chatID(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
		return update.CallbackQuery.Message.Message.Chat.ID
	}
	return 0
}
