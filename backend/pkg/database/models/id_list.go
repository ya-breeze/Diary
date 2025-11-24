package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

//nolint:recvcheck // it should impelement Valuer by value for some reasons
type StringList []string

// Implement the sql.Scanner interface
func (s *StringList) Scan(value any) error {
	str, ok := value.(string)
	if ok {
		if len(str) == 0 {
			return nil
		}

		if str[0] != '[' {
			*s = strings.Split(str, ",")
			return nil
		}

		return json.Unmarshal([]byte(str), s)
	}

	bytes, ok := value.([]byte)
	if ok {
		return json.Unmarshal(bytes, s)
	}

	return fmt.Errorf("failed to unmarshal StringList value '%+v'", value)
}

// Implement the driver.Valuer interface
func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}
