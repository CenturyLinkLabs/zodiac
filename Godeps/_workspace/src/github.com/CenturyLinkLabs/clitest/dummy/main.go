package main

import (
	"flag"
	"fmt"
	"os"
)

var flagPassed bool
var shouldExplode bool

func init() {
	flag.BoolVar(&flagPassed, "test", false, "used for testing passed arguments")
	flag.BoolVar(&shouldExplode, "explode", false, "used to request bad exit")
}

func main() {
	flag.Parse()

	if flagPassed {
		fmt.Print("You set the test flag")
	} else if shouldExplode {
		fmt.Fprintf(os.Stderr, "I exploded")
		os.Exit(19)
	} else if v := os.Getenv("CLITEST_TEST_VAR"); v != "" {
		fmt.Printf("CLITEST_TEST_VAR is %s", v)
	} else {
		fmt.Print("No Arguments Passed")
	}
}
