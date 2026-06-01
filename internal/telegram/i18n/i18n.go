// Package i18n provides UI strings for the Telegram bot (MVP: ru + en).
package i18n

import "strings"

// LocaleFromTelegram maps Telegram language_code to app locale.
func LocaleFromTelegram(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if strings.HasPrefix(code, "ru") {
		return "ru"
	}
	return "en"
}

var messages = map[string]map[string]string{
	"welcome": {
		"ru": "Привет! Я EasyTerms — помогу разобрать пользовательское соглашение.\n\nСоздайте новый документ и отправьте текст или ссылку. Когда закончите — нажмите «Готово к разбору».",
		"en": "Hi! I'm EasyTerms — I help you understand terms of service.\n\nCreate a new document and send text or a URL. When ready, tap «Ready to analyze».",
	},
	"new_doc_created": {
		"ru": "Новый документ создан. Отправьте текст или ссылку на страницу соглашения.",
		"en": "New document created. Send text or a URL to the agreement page.",
	},
	"no_active_doc": {
		"ru": "Сначала создайте новый документ.",
		"en": "Create a new document first.",
	},
	"content_added": {
		"ru": "Добавлено. Можно отправить ещё текст или ссылку, либо нажать «Готово к разбору».",
		"en": "Added. You can send more text or a URL, or tap «Ready to analyze».",
	},
	"ingest_ok": {
		"ru": "Документ готов. Выберите режим анализа:",
		"en": "Document is ready. Choose an analysis mode:",
	},
	"insufficient_balance": {
		"ru": "Недостаточно проверок. Купите пакет или обратитесь к администратору.",
		"en": "Not enough checks left. Buy a package or contact admin.",
	},
	"balance": {
		"ru": "Баланс проверок: %d",
		"en": "Check balance: %d",
	},
	"error_generic": {
		"ru": "Произошла ошибка. Попробуйте позже.",
		"en": "Something went wrong. Please try again later.",
	},
	"err_ingest_failed": {
		"ru": "Не удалось обработать документ. Проверьте текст или ссылку и попробуйте снова.",
		"en": "Could not process the document. Check the text or URL and try again.",
	},
	"err_url_fetch": {
		"ru": "Не удалось загрузить страницу по ссылке. Проверьте URL или вставьте текст вручную.",
		"en": "Could not load the page from the URL. Check the link or paste the text manually.",
	},
	"disclaimer": {
		"ru": "⚠️ Это не юридическая консультация. Решение о принятии условий остаётся за вами.",
		"en": "⚠️ This is not legal advice. You decide whether to accept the terms.",
	},
	"demo": {
		"ru": "Пример (демо, без списания проверок):\n\n«Сервис может изменять условия без уведомления. Продолжая использование, вы соглашаетесь с обновлениями. Персональные данные обрабатываются для работы сервиса и могут передаваться партнёрам.»\n\nСоздайте новый документ и вставьте свой текст или ссылку для реального разбора.",
		"en": "Example (demo, no checks charged):\n\n\"The service may change terms without notice. By continuing to use it, you accept updates. Personal data is processed to operate the service and may be shared with partners.\"\n\nCreate a new document and paste your own text or URL for a real analysis.",
	},
	"buy_intro": {
		"ru": "Пакеты проверок (MVP — оплата вручную через администратора):",
		"en": "Check packages (MVP — manual payment via admin):",
	},
	"buy_manual": {
		"ru": "Выбран пакет %s. Напишите администратору для зачисления проверок (ID покупки: %s).",
		"en": "Package %s selected. Contact admin to credit checks (purchase ID: %s).",
	},
	"analysis_plain": {
		"ru": "Простое объяснение:\n\n%s",
		"en": "Plain summary:\n\n%s",
	},
	"analysis_highlights": {
		"ru": "Важные пункты:\n\n%s",
		"en": "Highlights:\n\n%s",
	},
	"btn_new_doc":       {"ru": "Новый документ", "en": "New document"},
	"btn_ready":         {"ru": "Готово к разбору", "en": "Ready to analyze"},
	"btn_plain":         {"ru": "Объяснить просто", "en": "Explain simply"},
	"btn_highlights":    {"ru": "Подсветить риски", "en": "Highlight risks"},
	"btn_balance":       {"ru": "Баланс", "en": "Balance"},
	"btn_buy":           {"ru": "Купить проверки", "en": "Buy checks"},
	"btn_demo":          {"ru": "Пример", "en": "Example"},
	"btn_pkg_1":         {"ru": "1 проверка", "en": "1 check"},
	"btn_pkg_3":         {"ru": "3 проверки", "en": "3 checks"},
	"btn_pkg_10":        {"ru": "10 проверок", "en": "10 checks"},
}

// T returns a localized string; falls back to English then the key.
func T(locale, key string) string {
	if m, ok := messages[key]; ok {
		if s, ok := m[locale]; ok {
			return s
		}
		if s, ok := m["en"]; ok {
			return s
		}
	}
	return key
}
