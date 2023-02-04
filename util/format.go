package util

import (
	"strings"
	"time"
)

func FormatCEP(cep string) string {
	if len(cep) < 8 {
		cep = strings.Repeat("0", 8-len(cep)) + cep
	}
	return cep[:5] + "-" + cep[5:8]
}

func FormatPhone(phone string) string {
	switch len(phone) {
	case 8:
		return phone[:4] + "-" + phone[4:]
	case 9:
		return phone[:5] + "-" + phone[5:]
	case 10:
		return "(" + phone[:2] + ") " + phone[2:6] + "-" + phone[6:]
	case 11:
		return "(" + phone[:2] + ") " + phone[2:7] + "-" + phone[7:]
	}
	return phone
}

func FormatTime(time *time.Time) string {
	if time == nil {
		return "?"
	}
	return time.Format("02/01/2006 15:04")
}

func FormatDate(time *time.Time) string {
	if time == nil {
		return "?"
	}
	return time.Format("02/01/2006")
}
