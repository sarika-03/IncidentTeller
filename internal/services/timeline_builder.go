package services

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
)

type TimelineBuilder struct{}

func NewTimelineBuilder() *TimelineBuilder {
	return &TimelineBuilder{}
}

func (t *TimelineBuilder) Generate(incident domain.Incident) (string, error) {
	var out string
	out += "🧠 Incident Timeline\n\n"

	for _, e := range incident.Events {
		out += fmt.Sprintf(
			"🕒 %s — %s (%s)\n",
			e.OccurredAt.Format(time.RFC3339),
			e.Description,
			e.Status,
		)
	}

	return out, nil
}
