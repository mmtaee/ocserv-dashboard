package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type CSVStringList []string

const (
	Free            = "Free"
	MonthlyTransmit = "MonthlyTransmit"
	MonthlyReceive  = "MonthlyReceive"
	TotallyTransmit = "TotallyTransmit"
	TotallyReceive  = "TotallyReceive"
)

func (s *CSVStringList) Value() (driver.Value, error) {
	return strings.Join(*s, ","), nil
}

func (s *CSVStringList) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("CSVStringList: failed to scan type %T", value)
	}
	if str == "" {
		*s = []string{}
	} else {
		*s = strings.Split(str, ",")
	}
	return nil
}

func (s *CSVStringList) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(*s))
}

func (s *CSVStringList) UnmarshalJSON(b []byte) error {
	var arr []string
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	*s = arr
	return nil
}
