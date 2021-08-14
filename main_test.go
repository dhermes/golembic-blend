package golembic_test

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	err := requireDB(ctx)
	if err != nil {
		// NOTE: It's somewhat moot to use `STDERR` in a Go test; see
		//       https://github.com/golang/go/issues/13976
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer poolClose()

	exitCode := m.Run()
	os.Exit(exitCode)
}
