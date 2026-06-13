package countline_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/KEINOS/go-countline/countline"
)

func ExampleCountLines() {
	for _, sample := range []struct {
		Input string
	}{
		{""},            // --> 0
		{"Hello"},       // --> 1
		{"Hello\n"},     // --> 1
		{"\n"},          // --> 1
		{"\n\n"},        // --> 2
		{"\nHello"},     // --> 2
		{"\nHello\n"},   // --> 2
		{"\n\nHello"},   // --> 3
		{"\n\nHello\n"}, // --> 3
	} {
		readerFile := strings.NewReader(sample.Input)

		count, err := countline.CountLines(readerFile)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%#v --> %v\n", sample.Input, count)
	}
	// Output:
	// "" --> 0
	// "Hello" --> 1
	// "Hello\n" --> 1
	// "\n" --> 1
	// "\n\n" --> 2
	// "\nHello" --> 2
	// "\nHello\n" --> 2
	// "\n\nHello" --> 3
	// "\n\nHello\n" --> 3
}
