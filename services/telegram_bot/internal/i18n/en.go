package i18n

// All values are sent to Telegram with parse_mode=HTML. Allowed tags:
//   <b>, <i>, <u>, <s>, <code>, <pre>, <a href="...">.
// User-supplied values interpolated via T(...) MUST be HTML-escaped by the
// caller (see handlers.htmlEscape) — this catalog is plain HTML otherwise.
var en = map[Key]string{
	Welcome: "👋 <b>Welcome to the Ocserv Dashboard bot.</b>\n\n" +
		"Here you can:\n" +
		"• 🔗 Link an existing VPN account\n" +
		"• 📊 Check usage and expiry\n" +
		"• 🔄 Request a renewal\n" +
		"• 🆕 Order a brand-new account\n" +
		"• 🔔 Get notified when your traffic runs low",

	BotDisabled:       "⚠️ The bot is currently disabled by the administrator. Please try again later.",
	MainMenu:          "📋 <b>Main menu</b>\nWhat would you like to do?",
	BtnAddAccount:     "🔗 Add Account",
	BtnMyAccounts:     "👤 My Accounts",
	BtnNewOrder:       "🆕 Order New Account",
	BtnHelp:           "ℹ️ Help",
	BtnLanguage:       "🌐 Language",
	BtnCancel:         "❌ Cancel",
	BtnBack:           "⬅️ Back",
	BtnUsage:          "📊 Usage",
	BtnRenew:          "🔄 Renew",
	BtnRemove:         "🗑 Remove",
	AskUsername:       "🔑 Please send your <b>VPN username</b>:",
	AskPassword:       "🔒 Now send your <b>VPN password</b> (it will be deleted from the chat right after):",
	AskUsernameNew:    "📝 Pick a username for your new account (3–32 chars, letters, digits, <code>_ - .</code>):",
	AskMessage:        "💬 Add an optional note for the admin, or send /skip to continue without one:",
	AskReceipt:        "🧾 Please send the payment receipt as a <b>photo</b>.",
	AuthSuccess:       "✅ Account linked successfully.",
	AuthFail:          "❌ Invalid username or password.",
	AuthLocked:        "🔒 Your account is locked. Please contact the administrator.",
	AlreadyLinked:     "ℹ️ This account is already linked to your Telegram chat.",
	NoAccounts:        "📭 You have not linked any accounts yet. Use <b>Add Account</b> from the main menu.",
	NoPackages:        "📦 No packages are available right now. Please try again later.",
	PickPackage:       "📦 <b>Pick a package:</b>",
	PickAccountRenew:  "🔄 Pick an account to renew:",
	RequestCreated:    "📨 Your request was submitted. The administrator will review it shortly.",
	RequestExists:     "⏳ You already have a pending request. Please wait until it is processed.",
	WaitForApproval:   "⏳ Waiting for admin approval…",
	NotApprovedYet:    "ℹ️ Your request is not approved yet — you cannot upload a receipt right now.",
	ReceiptSaved:      "🧾 Receipt received. Awaiting admin confirmation.",
	OnlyPhoto:         "📷 Please send the receipt as a photo.",
	HelpText: "<b>Ocserv Dashboard bot — help</b>\n\n" +
		"Commands:\n" +
		"• /start — open the main menu\n" +
		"• /help — show this help\n" +
		"• /settings — bot settings\n" +
		"• /language — change language\n" +
		"• /cancel — cancel the current operation\n\n" +
		"Use the inline buttons to manage your VPN accounts, view usage, request renewals, order new accounts and upload payment receipts.",
	UsageText: "👤 <b>Account:</b> <code>%s</code>\n" +
		"📌 <b>Status:</b> %s\n" +
		"💾 <b>Quota:</b> %d GB\n" +
		"⬇️ <b>Used RX:</b> %.2f GB\n" +
		"⬆️ <b>Used TX:</b> %.2f GB\n" +
		"📅 <b>Expires:</b> %s",
	AccountRemoved:    "🗑 Account unlinked from your Telegram chat.",
	NotLinked:         "❓ Account is not linked to your Telegram chat.",
	UnknownCommand:    "🤔 Unknown command. Use the menu buttons.",
	LowQuotaWarning:   "🔔 <b>Low quota warning for</b> <code>%s</code>: only %d MB remaining. Please consider renewing.",
	LanguagePicked:    "✅ Language updated.",
	SessionTimedOut:   "⌛ Session timed out. Please try again from the main menu.",
	OcservDeactivated: "⛔ This account is deactivated.",
	RateLimited:       "🚦 Too many attempts. Please wait a minute and try again.",
}
