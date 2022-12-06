package util

import (
	"os/exec"
	"runtime"
)

func GetWkhtmlPdfCmd(args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("wkhtmltopdf.exe", args...)
	} else {
		return exec.Command("wkhtmltopdf", args...)
	}
}
