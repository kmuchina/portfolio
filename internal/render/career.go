package render

import (
	"fmt"
	"time"

	"github.com/kmuchina/portfolio/internal/career"
	"github.com/kmuchina/portfolio/internal/site"
)

// careerGraph converts the experience items and education milestones into
// career entries and builds the graph for "now". Validation has already
// checked the dates; an error here means a bug, not bad content.
func careerGraph(s *site.Site, now time.Time) (*career.Graph, error) {
	var entries []career.Entry
	for i, j := range s.Experience.Items {
		start, err := career.ParseMonth(j.Start)
		if err != nil {
			return nil, fmt.Errorf("experience.items[%d].start: %w", i, err)
		}
		e := career.Entry{
			ID: j.ID(), Title: j.Title, Org: j.Org, Type: j.Type, Icon: j.Icon, Location: j.Location,
			Start: start, Present: j.End == site.Present, Bullets: j.Bullets, Skills: j.Skills,
		}
		if !e.Present {
			if e.End, err = career.ParseMonth(j.End); err != nil {
				return nil, fmt.Errorf("experience.items[%d].end: %w", i, err)
			}
		}
		entries = append(entries, e)
	}
	for i, school := range s.Education.Items {
		m := school.Timeline
		if m == nil {
			continue
		}
		date, err := career.ParseMonth(m.Date)
		if err != nil {
			return nil, fmt.Errorf("education.items[%d].timeline.date: %w", i, err)
		}
		entries = append(entries, career.Entry{
			Title: m.Title, Org: m.Detail, Icon: "graduation-cap",
			Start: date, End: date, Milestone: true,
		})
	}
	ui := s.UI.Career
	g := career.Build(entries, career.MonthOf(now), career.Labels{
		All: ui.All, Milestone: ui.Milestone, Present: ui.Present,
		ShowingAll: ui.ShowingAll, Showing: ui.Showing, ShowingMilestones: ui.ShowingMilestones,
	})
	return &g, nil
}
