package stats

type UserStats struct {
	Username  string
	IP        string
	SessionID string
	RX        int
	TX        int
}

type Totals struct {
	TotalRx int
	TotalTx int
}
