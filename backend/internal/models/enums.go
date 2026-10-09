package models

// ExpiryMode defines how an Ocserv user's effective expiry is determined.
type ExpiryMode string

const (
	ExpiryModeUnlimited       ExpiryMode = "unlimited"
	ExpiryModeFixed           ExpiryMode = "fixed"
	ExpiryModeFirstConnection ExpiryMode = "first_connection"
)

func (mode ExpiryMode) IsValid() bool {
	switch mode {
	case ExpiryModeUnlimited, ExpiryModeFixed, ExpiryModeFirstConnection:
		return true
	default:
		return false
	}
}

// TrafficType defines how an Ocserv user's traffic quota is measured.
type TrafficType string

const (
	Free            TrafficType = "Free"
	MonthlyTransmit TrafficType = "MonthlyTransmit"
	MonthlyReceive  TrafficType = "MonthlyReceive"
	MonthlyRxTx     TrafficType = "MonthlyRxTx"
	TotallyTransmit TrafficType = "TotallyTransmit"
	TotallyReceive  TrafficType = "TotallyReceive"
	TotallyRxTx     TrafficType = "TotallyRxTx"
)

func (trafficType TrafficType) IsValid() bool {
	switch trafficType {
	case Free, MonthlyTransmit, MonthlyReceive, MonthlyRxTx, TotallyTransmit, TotallyReceive, TotallyRxTx:
		return true
	default:
		return false
	}
}

// OcservUserSessionEvent identifies an event emitted by the Ocserv log stream.
type OcservUserSessionEvent string

const (
	EventUseragent     OcservUserSessionEvent = "user-agent"
	EventHandshake     OcservUserSessionEvent = "handshake"
	EventPeriodicStats OcservUserSessionEvent = "periodic-stats"
	EventDisconnect    OcservUserSessionEvent = "disconnect"
)

func (event OcservUserSessionEvent) IsValid() bool {
	switch event {
	case EventUseragent, EventHandshake, EventPeriodicStats, EventDisconnect:
		return true
	default:
		return false
	}
}

// OcservUserCertificateStatus identifies whether a backed-up certificate is usable.
type OcservUserCertificateStatus string

const (
	OcservUserCertificateStatusActive    OcservUserCertificateStatus = "active"
	OcservUserCertificateStatusSuspended OcservUserCertificateStatus = "suspended"
)

func (status OcservUserCertificateStatus) IsValid() bool {
	switch status {
	case OcservUserCertificateStatusActive, OcservUserCertificateStatusSuspended:
		return true
	default:
		return false
	}
}

// TelegramLanguage is a supported language code for Telegram UI content.
type TelegramLanguage string

const (
	TelegramLanguageEN   TelegramLanguage = "en"
	TelegramLanguageFA   TelegramLanguage = "fa"
	TelegramLanguageAR   TelegramLanguage = "ar"
	TelegramLanguageRU   TelegramLanguage = "ru"
	TelegramLanguageZHCN TelegramLanguage = "zh-cn"
	TelegramLanguageZHTW TelegramLanguage = "zh-tw"
	TelegramLanguageIT   TelegramLanguage = "it"
)

func (language TelegramLanguage) IsValid() bool {
	switch language {
	case TelegramLanguageEN, TelegramLanguageFA, TelegramLanguageAR, TelegramLanguageRU,
		TelegramLanguageZHCN, TelegramLanguageZHTW, TelegramLanguageIT:
		return true
	default:
		return false
	}
}

// TelegramRequestType identifies the operation requested through the bot.
type TelegramRequestType string

const (
	TelegramRequestTypeNew   TelegramRequestType = "new"
	TelegramRequestTypeRenew TelegramRequestType = "renew"
)

func (requestType TelegramRequestType) IsValid() bool {
	switch requestType {
	case TelegramRequestTypeNew, TelegramRequestTypeRenew:
		return true
	default:
		return false
	}
}

// TelegramRequestStatus identifies the lifecycle state of a Telegram request.
type TelegramRequestStatus string

const (
	TelegramRequestStatusPending         TelegramRequestStatus = "pending"
	TelegramRequestStatusAwaitingPayment TelegramRequestStatus = "awaiting_payment"
	TelegramRequestStatusPaymentUploaded TelegramRequestStatus = "payment_uploaded"
	TelegramRequestStatusApproved        TelegramRequestStatus = "approved"
	TelegramRequestStatusRejected        TelegramRequestStatus = "rejected"
	TelegramRequestStatusDelivered       TelegramRequestStatus = "delivered"
)

func (status TelegramRequestStatus) IsValid() bool {
	switch status {
	case TelegramRequestStatusPending, TelegramRequestStatusAwaitingPayment,
		TelegramRequestStatusPaymentUploaded, TelegramRequestStatusApproved,
		TelegramRequestStatusRejected, TelegramRequestStatusDelivered:
		return true
	default:
		return false
	}
}
