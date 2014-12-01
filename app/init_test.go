package app

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/campadrenalin/contentgremlin/cgdb"
	gtc "github.com/campadrenalin/go-test-common"
	"github.com/stretchr/testify/assert"
)

func TestInitDirectory(t *testing.T) {
	td_tracker := gtc.NewTempDirTracker(t)
	defer td_tracker.Cleanup()

	tests := []struct {
		dir             string
		expected_output string
		error_text      string
	}{
		{
			dir: td_tracker.Create(),
			expected_output: `init_test: Attempting to initialize in TEMPDIR...
init_test: Successfully initialized CG dir in TEMPDIR
`,
			error_text: "",
		},
		{
			dir: td_tracker.Create() + "/foo",
			expected_output: `init_test: Attempting to initialize in TEMPDIR...
init_test: Successfully initialized CG dir in TEMPDIR
`,
			error_text: "",
		},
	}

	for _, test := range tests {
		output := gtc.NewWriteCompare()
		logger := log.New(output, "init_test: ", 0)

		err := InitDirectory(test.dir, logger)
		gtc.AssertError(t, err, test.error_text, false)
		if err != nil {
			// Skip to next, whether error was intentional or not
			continue
		}

		// Nothing funky about contents dir
		if _, err := os.Stat(path.Join(test.dir, "contents")); err != nil {
			t.Fatal(err)
		}

		db, err := cgdb.Open(test.dir)
		if err != nil {
			t.Fatal(err)
		}
		version, err := db.GetVersion()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, version)

		output.Replacements[test.dir] = "TEMPDIR"
		assert.Equal(t, test.expected_output, output.String())
	}
}
