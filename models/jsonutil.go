package models

import "fmt"

func fmtString(s fmt.Stringer) string {
	if s != nil {
		return s.String()
	}
	return ""
}
