package telegram

// Allowlist gates bot access by Telegram user ID. Empty slice means allow everyone.
type Allowlist struct {
	ids map[int64]struct{}
}

// NewAllowlist builds an allowlist. When ids is empty, all users are allowed.
func NewAllowlist(ids []int64) Allowlist {
	if len(ids) == 0 {
		return Allowlist{}
	}
	m := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return Allowlist{ids: m}
}

// Open reports whether any Telegram user may use the bot.
func (a Allowlist) Open() bool {
	return len(a.ids) == 0
}

// Allowed reports whether telegramID may use the bot.
func (a Allowlist) Allowed(telegramID int64) bool {
	if a.Open() {
		return true
	}
	_, ok := a.ids[telegramID]
	return ok
}
