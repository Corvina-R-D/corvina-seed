package utils

import "fmt"

var colorReset = "\033[0m"
var colorRed = "\033[31m"
var colorGreen = "\033[32m"
var colorYellow = "\033[33m"
var colorBlue = "\033[34m"
var colorPurple = "\033[35m"
var colorCyan = "\033[36m"
var colorWhite = "\033[37m"

func PrintlnGreen(message string) {
	fmt.Println(colorGreen + message + colorReset)
}

func PrintlnRed(message string) {
	fmt.Println(colorRed + message + colorReset)
}

func PrintlnYellow(message string) {
	fmt.Println(colorYellow + message + colorReset)
}

func PrintlnBlue(message string) {
	fmt.Println(colorBlue + message + colorReset)
}

func PrintlnPurple(message string) {
	fmt.Println(colorPurple + message + colorReset)
}

func PrintlnCyan(message string) {
	fmt.Println(colorCyan + message + colorReset)
}

func PrintlnWhite(message string) {
	fmt.Println(colorWhite + message + colorReset)
}
