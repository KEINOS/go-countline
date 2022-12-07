package spec

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/KEINOS/go-countline/cl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStrDummyLines(t *testing.T) {
	t.Parallel()

	const sizeLine = bufio.MaxScanTokenSize * 2

	for _, numLine := range []int64{
		1,
		2,
	} {
		result := GetStrDummyLines(sizeLine, numLine)
		require.NotEmpty(t, result, "the result should not be empty")

		{
			expectLen := (65536 * 2) * numLine // bufio.MaxScanTokenSize * <num of lines> + <num of line breaks>
			actualLen := int64(len(result))

			require.Equal(t, expectLen, actualLen, "byte length miss match. numLine: %v", numLine)
		}
		{
			expectCount := numLine
			actualCount := int64(strings.Count(result, "\n"))

			require.Equal(t, expectCount, actualCount, "it should have %v line breaks", numLine)
		}
		{
			expectHead := "a"
			actualHead := string(result[0])

			require.Equal(t, expectHead, actualHead, "the first character should be 'a'")
		}
		{
			expectEnd := byte('\n')
			actualEnd := result[len(result)-1]

			require.Equal(t, expectEnd, actualEnd, "the last character should be a line break")
		}
	}
}

func TestRunSpecTest(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() {
		RunSpecTest(t, "CountLines", cl.CountLines)
	})
}

func Test_genOneLine(t *testing.T) {
	t.Parallel()

	{
		dataLine := genOneLine(0)
		require.Empty(t, dataLine, "line size 0 should be empty")
	}

	for index, test := range []struct {
		name        string
		expectFirst byte
		expectLast  byte
		requestSize int64
	}{
		{
			name:        "line size of 1 should have length of 1 and contains only \\n",
			requestSize: 1,
			expectFirst: byte(0x0a),
			expectLast:  byte(0x0a),
		},
		{
			name:        "line size of 2 should have length of 2 and contains only 'a\\n'",
			requestSize: 2,
			expectFirst: byte('a'),
			expectLast:  byte(0x0a),
		},
		{
			name:        "line size of 100 should have length of 100 and contains only 'a' and the last char as '\\n'",
			requestSize: 100,
			expectFirst: byte('a'),
			expectLast:  byte(0x0a),
		},
	} {
		dataLine := genOneLine(test.requestSize)
		t.Logf("Generated data: %#v", dataLine)

		reason := fmt.Sprintf("test %d: %s", index+1, test.name)

		require.Equal(t, int64(len(dataLine)), test.requestSize, reason)

		assert.Equal(t, dataLine[0], test.expectFirst, "first char did not match.\n%s", reason)

		for i := 0; i < len(dataLine)-1; i++ {
			assert.Equal(t, dataLine[i], byte('a'), "index %d should be 'a'\n%s", i, reason)
		}

		assert.Equal(t, dataLine[len(dataLine)-1], test.expectLast, "last char did not match. %s", reason)
	}
}
