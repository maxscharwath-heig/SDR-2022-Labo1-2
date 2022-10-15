package utils

import (
	"fmt"
	"sdr/labo1/utils/colors"
	"time"
)

var enabled = true

func SetEnabled(enable bool) {
	enabled = enable
}

func LogInfo(prefix string, data ...any) {
	Log(prefix, colors.Yellow, data)
}

func LogError(data ...any) {
	Log("error", colors.Red, data)
}

func Log(prefix string, color string, data ...any) {
	if !enabled {
		return
	}
	date := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(color, fmt.Sprintf("[%s] (%s):", date, prefix), colors.Reset, data)
}
