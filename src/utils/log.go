package utils

import (
	"fmt"
	"sdr/labo1/src/utils/colors"
	"time"
)

var enabled = true

func SetEnabled(enable bool) {
	enabled = enable
}

func LogInfo(prefix string, data ...any) {
	Log(false, prefix, colors.Yellow, data)
}

func LogSuccess(prefix string, data ...any) {
	Log(false, prefix, colors.Green, data)
}

func LogError(data ...any) {
	Log(false, "error", colors.Red, data)
}

func Log(force bool, prefix string, color string, data ...any) {
	if !enabled && !force {
		return
	}
	date := time.Now().Format("2006-01-02 15:04:05")
	var result []any
	result = append(result, color+fmt.Sprintf("[%s] (%s):", date, prefix)+colors.Reset)
	result = append(result, data...)
	fmt.Println(result...)
}