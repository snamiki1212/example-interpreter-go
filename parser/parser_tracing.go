package parser

import (
	"fmt"
	"strings"
)

var traceLevel int

const traceIdentPlaceholder string = "\t"

const IS_DEBUG = false

func identLevel() string {
	return strings.Repeat(traceIdentPlaceholder, traceLevel-1)
}

func tracePrint(fs string) {
	if IS_DEBUG {
		fmt.Printf("%s%s\n", identLevel(), fs)
	}
}

func incIdent() { traceLevel = traceLevel + 1 }
func decIdent() { traceLevel = traceLevel - 1 }

func trace(msg string) string {
	incIdent()
	tracePrint("BEGIN " + msg)
	return msg
}

func untrace(msg string) {
	tracePrint("END " + msg)
	decIdent()
}
