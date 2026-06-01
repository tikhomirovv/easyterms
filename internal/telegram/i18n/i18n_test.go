package i18n_test

import (
	"testing"

	"github.com/tikhomirovv/easyterms/internal/telegram/i18n"
)

func TestLocaleFromTelegram(t *testing.T) {
	if i18n.LocaleFromTelegram("ru-RU") != "ru" {
		t.Fatal("expected ru")
	}
	if i18n.LocaleFromTelegram("en") != "en" {
		t.Fatal("expected en")
	}
}

func TestT_fallback(t *testing.T) {
	s := i18n.T("ru", "welcome")
	if s == "" || s == "welcome" {
		t.Fatalf("got %q", s)
	}
}
