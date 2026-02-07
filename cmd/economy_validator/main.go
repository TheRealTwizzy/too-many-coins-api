package main

import (
	"flag"
	"fmt"
)

func main() {
	extended := flag.Bool("extended", true, "Run extended validations with longer durations")
	flag.Parse()

	if *extended {
		fmt.Println("\n" + repeatStr("=", 80))
		fmt.Println("🔍 ECONOMY POPULATION-INVARIANCE VALIDATION")
		fmt.Println("Extended Mode: Long-Duration Scenarios")
		fmt.Println(repeatStr("=", 80))
		RunExtendedValidationScenarios()
	}
}

func repeatStr(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
