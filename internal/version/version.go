// Package version holds the application release version (updated on each release).
package version

// Version is the current release. Bumped by the release skill before tagging.
// Logged at bot startup (see cmd/telegram); not shown in Telegram UI unless we add a /version command later.
const Version = "0.1.1"
