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

type PlanStatus string

const (
	PlanStatusUnset   PlanStatus = ""
	PlanStatusFailed  PlanStatus = migration.StatFailed
	PlanStatusApplied PlanStatus = migration.StatApplied
)

type PlanEvent struct {
	Revision string
	Body     string
	Status   PlanStatus
	Labels   []string
}

func (PlanEvent) GetFlag() string {
	return migration.Flag
}

func (pe PlanEvent) Result() string {
	if pe.Revision == "" {
		return "plan"
	}
	return pe.Revision
}

func (pe PlanEvent) Color() ansi.Color {
	if pe.Status == PlanStatusApplied {
		return ansi.ColorBlue
	}
	if pe.Status == PlanStatusFailed {
		return ansi.ColorRed
	}
	return ansi.ColorGreen
}

// WriteText writes the migration event as text.
func (pe PlanEvent) WriteText(tf logger.TextFormatter, wr io.Writer) {
	fmt.Fprint(wr, tf.Colorize("--", ansi.ColorLightBlack))
	fmt.Fprint(wr, logger.Space)
	fmt.Fprint(wr, tf.Colorize(pe.Result(), pe.Color()))

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
	m := map[string]interface{}{
		"labels": pe.Labels,
		"body":   pe.Body,
	}
	if pe.Revision == "" {
		m["result"] = "plan"
	} else {
		m["revision"] = pe.Revision
	}
	if pe.Status != "" {
		m["status"] = pe.Status
	}
	return m
}

func PlanEventWrite(ctx context.Context, log logger.Log, revision, body string, status PlanStatus) {
	pe := PlanEvent{Revision: revision, Body: body, Status: status, Labels: migration.GetContextLabels(ctx)}
	logger.MaybeTriggerContext(ctx, log, pe)
}
