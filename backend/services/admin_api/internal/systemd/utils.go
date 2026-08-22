package systemd

import (
	"strconv"
	"strings"
)

func ParseSystemctlShow(output string) OcservSystemdStatus {
	lines := strings.Split(output, "\n")

	data := make(map[string]string)

	for _, line := range lines {
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		data[kv[0]] = kv[1]
	}

	return OcservSystemdStatus{
		ID:            data["Id"],
		Description:   data["Description"],
		ActiveState:   data["ActiveState"],
		SubState:      data["SubState"],
		UnitFileState: data["UnitFileState"],

		MainPID: toInt(data["MainPID"]),
		//StartTime:    parseTime(data["ExecMainStartTimestamp"]),
		StartTime:    data["ExecMainStartTimestamp"],
		Memory:       toInt64(data["MemoryCurrent"]),
		CPUUsageNSec: toInt64(data["CPUUsageNSec"]),
		Tasks:        toInt(data["TasksCurrent"]),
	}
}

func toInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func toInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

//func parseTime(s string) time.Time {
//	// systemd format: Fri 2026-04-24 14:35:45 +0330
//	layout := "Mon 2006-01-02 15:04:05 -0700"
//	t, _ := time.Parse(layout, s)
//	return t
//}
