// Package cbdata holds Telegram inline_keyboard callback_data strings.
// Router and handlers must use these constants only so mismatches are compile-time errors.
package cbdata

const (
	MainMenu         = "menu:main"
	AddAccount       = "menu:add"
	MyAccounts       = "menu:list"
	NewOrder         = "menu:order"
	Help             = "menu:help"
	Language         = "menu:lang"
	AccountDetail    = "acc:detail:"
	AccountUsage     = "acc:usage:"
	AccountRenew     = "acc:renew:"
	AccountRemove    = "acc:remove:"
	PickPackageNew   = "pkgn:"
	PickPackageRenew = "pkgr:"

	AdminMenu     = "adm:menu"
	AdminPending  = "adm:pending"
	AdminReceipts = "adm:receipts"
	AdminStats    = "adm:stats"
)
