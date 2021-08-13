package golembic

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/blend/go-sdk/ansi"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/logger"
)

// NOTE: Ensure that
//       * `PlanEvent` satisfies `logger.Event`.
//       * `PlanEvent` satisfies `logger.TextWritable`.
//       * `PlanEvent` satisfies `logger.JSONWritable`.
var (
	_ logger.Event        = (*PlanEvent)(nil)
	_ logger.TextWritable = (*PlanEvent)(nil)
	_ logger.JSONWritable = (*PlanEvent)(nil)
)

type PlanEvent struct {
	Body   string
	Labels []string
}

func (PlanEvent) GetFlag() string {
	return migration.Flag
}

// WriteText writes the migration event as text.
func (pe PlanEvent) WriteText(tf logger.TextFormatter, wr io.Writer) {
	fmt.Fprint(wr, tf.Colorize("--", ansi.ColorLightBlack))
	fmt.Fprint(wr, logger.Space)
	fmt.Fprint(wr, tf.Colorize("plan", ansi.ColorGreen))

	if len(pe.Labels) > 0 {
		fmt.Fprint(wr, logger.Space)
		fmt.Fprint(wr, strings.Join(pe.Labels, " > "))
	}

	if len(pe.Body) > 0 {
		fmt.Fprint(wr, logger.Space)
		fmt.Fprint(wr, tf.Colorize("--", ansi.ColorLightBlack))
		fmt.Fprint(wr, logger.Space)
		fmt.Fprint(wr, pe.Body)
	}
}

// Decompose implements logger.JSONWritable.
func (pe PlanEvent) Decompose() map[string]interface{} {
	return map[string]interface{}{
		"labels": pe.Labels,
		"body":   pe.Body,
	}
}

func PlanEventWrite(ctx context.Context, log logger.Log, body string) {
	e := PlanEvent{Body: body, Labels: migration.GetContextLabels(ctx)}
	logger.MaybeTriggerContext(ctx, log, e)
}