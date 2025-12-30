/*
Package spec provides the test specifications for the CountLines function.

Alternate implementations of CountLines function must pass the test as well.
*/
//nolint:gochecknoglobals // allow global variable.
package spec

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
)

// ============================================================================
//  Data Provider of CountLines Specification
// ============================================================================

// DataCountLines is the data provider for the CountLines function to check if
// the specifications are covered.
// Alternate functions must pass the test with this data as well.
//
//nolint:mnd // numbers of ExpectOut are not magic numbers and let DataCountLines be global.
var DataCountLines = []struct {
	Reason    string // Reason on failure
	Input     string // Input data
	ExpectOut int    // Expected output
}{
	{
		Reason:    "'<EOF>' --> empty file should be zero",
		Input:     "",
		ExpectOut: 0,
	},
	{
		Reason:    "'Hello<EOF>' --> single line without line break should be one",
		Input:     "Hello",
		ExpectOut: 1,
	},
	{
		Reason:    "'Hello\\n<EOF>' --> single line with line break should be one",
		Input:     "Hello\n",
		ExpectOut: 1,
	},
	{
		Reason:    "'\\n<EOF>' --> single line break should be one",
		Input:     "\n",
		ExpectOut: 1,
	},
	{
		Reason:    "'\\n\\n<EOF>' --> two line breaks should be two",
		Input:     "\n\n",
		ExpectOut: 2,
	},
	{
		Reason:    "'\\nHello<EOF>' --> one line break and one line without line break should be two",
		Input:     "\nHello",
		ExpectOut: 2,
	},
	{
		Reason:    "'\\nHello\\n<EOF>' --> one line break and one line with line break should be two",
		Input:     "\nHello\n",
		ExpectOut: 2,
	},
	{
		Reason:    "'\\n\\nHello<EOF>' --> two line breaks and one line without line break should be three",
		Input:     "\n\nHello",
		ExpectOut: 3,
	},
	{
		Reason:    "'\\n\\nHello\\n<EOF>' --> two line breaks and one line with line break should be three",
		Input:     "\n\nHello\n",
		ExpectOut: 3,
	},
	{
		Reason:    "'<large line>\\n<EOF>' --> long string with a line break should be onw",
		Input:     GetStrDummyLines(bufio.MaxScanTokenSize*2, 1),
		ExpectOut: 1,
	},
	{
		Reason:    "'<large line>\\n<large line>\\n<EOF>' --> long string but two lines should be two",
		Input:     GetStrDummyLines(bufio.MaxScanTokenSize*2, 2),
		ExpectOut: 2,
	},
}

// ============================================================================
//  Functions
// ============================================================================

// ----------------------------------------------------------------------------
//  RunSpecTest
// ----------------------------------------------------------------------------

// RunSpecTest is a helper function to run the specifcations of LineCount function.
// Alternate implementations (_alt.*) must pass this test as well.
//
//nolint:varnamelen // fn is short for the scope of its usage but leave it as is.
func RunSpecTest(t *testing.T, nameFn string, fn func(io.Reader) (int, error)) {
	t.Helper()

	const threshold = 1024 // Max size of input data to begin cropping

	for index, test := range DataCountLines {
		testNum := fmt.Sprintf("test #%v", index)

		t.Run(testNum, func(t *testing.T) {
			logInput := test.Input

			// Crop the input to make it readable
			if len(test.Input) > threshold {
				logInput = fmt.Sprintf("%v ... %v", test.Input[:64], test.Input[len(test.Input)-64:])
			}

			t.Logf("Input : %#v", logInput)
			t.Log("Input len:", len(test.Input))

			ioReader := strings.NewReader(test.Input)

			msgReason := fmt.Sprintf("%v %v: %v", nameFn, testNum, test.Reason)
			expect := test.ExpectOut

			// Run the target function. Capture the STDOUT and STDERR as well.
			out := capturer.CaptureOutput(func() {
				actual, err := fn(ioReader)

				require.NoError(t, err, "golden case should not return error")
				assert.Equal(t, expect, actual, msgReason)
			})

			assert.Emptyf(t, out,
				"%v %v: it should not output to STDOUT/STDERR on success.\nOut: %v",
				nameFn, testNum, out,
			)
		})
	}
}

// ----------------------------------------------------------------------------
//  GetStrDummyLines
// ----------------------------------------------------------------------------

// GetStrDummyLines returns a string with the given size and number of lines.
// 'sizeLine' is the size of each line. 'numLines' is the number of lines.
func GetStrDummyLines(sizeLine, numLine int64) string {
	dataLine := genOneLine(sizeLine)

	// result := ""
	// for range numLine {
	// 	result += string(dataLine)
	// }
	//
	// return result

	var strBldr strings.Builder

	for range numLine {
		strBldr.Write(dataLine)
	}

	return strBldr.String()
}

// ----------------------------------------------------------------------------
//  genOneLine
// ----------------------------------------------------------------------------

func genOneLine(sizeLine int64) []byte {
	if sizeLine <= 0 {
		return []byte{}
	}

	const LF = byte(0x0a) //nolint:varnamelen

	dataLine := make([]byte, sizeLine)

	for i := range sizeLine {
		dataLine[i] = 'a'

		if i == sizeLine-1 {
			dataLine[i] = LF // \n
		}
	}

	return dataLine
}
