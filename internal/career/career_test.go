package career

import (
	"strconv"
	"strings"
	"testing"
)

var labels = Labels{
	All: "All roles", Milestone: "Milestone", Present: "Present",
	ShowingAll: "Showing all roles", Showing: "Showing {kind} roles", ShowingMilestones: "Showing milestones",
}

func m(s string) Month {
	v, err := ParseMonth(s)
	if err != nil {
		panic(err)
	}
	return v
}

func role(title, typ, start, end string) Entry {
	e := Entry{Title: title, Type: typ, Start: m(start)}
	if end == "present" {
		e.Present = true
	} else {
		e.End = m(end)
	}
	return e
}

func milestone(title, date string) Entry {
	return Entry{Title: title, Type: "Milestone", Start: m(date), End: m(date), Milestone: true}
}

// sketch renders one row as "l0[...] l1[...]" so tables read like the graph.
func sketch(r Row) string {
	var lanes [2][]string
	style := func(s Seg) string {
		switch {
		case s.Fade:
			return "fade"
		case s.Dashed:
			return "dashed"
		}
		return "solid"
	}
	for _, s := range r.Above {
		lanes[s.Lane] = append(lanes[s.Lane], s.Pos+":"+style(s))
	}
	if r.Curve != nil {
		kind := "merge"
		if r.Curve.Fork {
			kind = "fork"
		}
		lanes[1] = append(lanes[1], "curve:"+kind)
	}
	shape := strings.TrimPrefix(strings.TrimPrefix(r.Node.Class(), "node node--l0"), "node node--l1")
	if shape == "" {
		shape = "circle"
	}
	lanes[r.Node.Lane] = append(lanes[r.Node.Lane], "node:"+strings.TrimPrefix(strings.TrimSpace(shape), "node--"))
	for _, s := range r.Below {
		lanes[s.Lane] = append(lanes[s.Lane], s.Pos+":"+style(s))
	}
	return "l0[" + strings.Join(lanes[0], " ") + "] l1[" + strings.Join(lanes[1], " ") + "]"
}

func check(t *testing.T, g Graph, want []string) {
	t.Helper()
	if len(g.Rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(g.Rows), len(want))
	}
	for i, r := range g.Rows {
		if got := r.Title + " " + sketch(r); got != want[i] {
			t.Errorf("row %d:\n got  %s\n want %s", i+1, got, want[i])
		}
	}
}

// realData mirrors content/site.yaml (experience + the graduation milestone),
// deliberately in a non-sorted order.
func realData() []Entry {
	return []Entry{
		role("KRB", "Internship", "2024-05", "2025-01"),
		role("KFC", "Internship", "2026-08", "present"),
		role("Tech4Good", "Volunteer", "2024-02", "2024-12"),
		role("Chair", "Volunteer", "2023-07", "2024-07"),
		milestone("Graduated", "2025-12"),
	}
}

func TestRealData(t *testing.T) {
	g := Build(realData(), m("2026-10"), labels)
	// The §4.3 table: rows top to bottom, per lane.
	check(t, g, []string{
		"KFC l0[above:fade node:current below:dashed] l1[]",
		"Graduated l0[above:dashed node:ring below:dashed] l1[]",
		"KRB l0[above:dashed node:circle] l1[curve:merge below:solid]",
		"Tech4Good l0[] l1[above:solid node:diamond below:solid]",
		"Chair l0[] l1[above:solid node:diamond]",
	})

	// The prototype's exact class lists.
	if got := g.Rows[0].Above[0].Class(); got != "seg seg--l0 seg--above seg--fade" {
		t.Errorf("KFC stub class = %q", got)
	}
	if got := g.Rows[2].Below[0].Class(); got != "seg seg--l1 seg--below" {
		t.Errorf("KRB lane-1 class = %q", got)
	}
	if got := g.Rows[2].Curve.Path(); got != "M20 0 C 20 13, 46 9, 46 22" {
		t.Errorf("merge path = %q", got)
	}

	var durations, widths, kinds []string
	for _, r := range g.Rows {
		durations = append(durations, r.Duration)
		widths = append(widths, strconv.Itoa(r.TenurePct))
		kinds = append(kinds, string(r.Kind))
	}
	if got := strings.Join(durations, "|"); got != "3 mos||9 mos|11 mos|1 yr 1 mo" {
		t.Errorf("durations = %s", got)
	}
	if got := strings.Join(widths, "|"); got != "23|0|69|85|100" {
		t.Errorf("tenure widths = %s", got)
	}
	if got := strings.Join(kinds, "|"); got != "internship|milestone|internship|volunteer|volunteer" {
		t.Errorf("kinds = %s", got)
	}
	if r := g.Rows[0]; r.StartISO != "2026-08" || r.StartText != "Aug 2026" || r.EndText != "Present" || r.EndISO != "" {
		t.Errorf("KFC dates = %+v", [4]string{r.StartISO, r.StartText, r.EndISO, r.EndText})
	}
	if r := g.Rows[2]; r.EndISO != "2025-01" || r.EndText != "Jan 2025" {
		t.Errorf("KRB end = %s %s", r.EndISO, r.EndText)
	}

	if s := g.Summary; s.Roles != 4 || s.Concurrent != 3 || s.Text != "roles since 2023, up to three at once" {
		t.Errorf("summary = %+v", s)
	}
	var legend []string
	for _, l := range g.Legend {
		legend = append(legend, l.Label+"="+strconv.Itoa(l.Count)+"/"+l.Lane)
	}
	if got := strings.Join(legend, " "); got != "All roles=5/all Internship=2/work Volunteer=2/volunteer Milestone=1/milestone" {
		t.Errorf("legend = %s", got)
	}
	var says []string
	for _, l := range g.Legend {
		says = append(says, l.Announce)
	}
	if got := strings.Join(says, "; "); got != "Showing all roles; Showing internship roles; Showing volunteer roles; Showing milestones" {
		t.Errorf("legend announcements = %s", got)
	}
}

func TestSynthetic(t *testing.T) {
	now := m("2026-10")
	tests := []struct {
		name    string
		entries []Entry
		want    []string
	}{
		{
			name: "no volunteer roles",
			entries: []Entry{
				role("B", "Full-time", "2023-01", "2024-06"),
				role("A", "Contract", "2021-01", "2022-03"),
			},
			want: []string{
				"B l0[node:circle below:dashed] l1[]",
				"A l0[above:dashed node:circle] l1[]",
			},
		},
		{
			name: "only volunteer roles",
			entries: []Entry{
				role("V2", "Volunteer", "2022-01", "2022-12"),
				role("V1", "Volunteer", "2020-01", "2021-12"),
			},
			want: []string{
				"V2 l0[] l1[node:diamond below:solid]", // V1 ran until the month before V2: continuous
				"V1 l0[] l1[above:solid node:diamond]",
			},
		},
		{
			name:    "single entry",
			entries: []Entry{role("Solo", "Freelance", "2025-03", "2025-09")},
			want:    []string{"Solo l0[node:circle] l1[]"},
		},
		{
			name: "volunteer lane forks out of an active lane 0",
			entries: []Entry{
				role("V", "Volunteer", "2021-06", "2021-12"),
				role("W", "Full-time", "2020-01", "2023-12"),
			},
			want: []string{
				"V l0[] l1[node:diamond below:solid]",
				"W l0[node:circle] l1[curve:fork]",
			},
		},
		{
			name: "two overlapping internships share lane 0",
			entries: []Entry{
				role("A", "Internship", "2022-01", "2022-12"),
				role("B", "Internship", "2022-06", "2023-03"),
			},
			want: []string{
				"B l0[node:circle below:solid] l1[]",
				"A l0[above:solid node:circle] l1[]",
			},
		},
		{
			name: "lane 1 ends after the newest lane-0 role started: merges into it",
			entries: []Entry{
				role("W", "Full-time", "2024-01", "present"),
				role("V", "Volunteer", "2023-01", "2024-06"),
			},
			want: []string{
				"W l0[above:fade node:current] l1[curve:merge below:solid]",
				"V l0[] l1[above:solid node:diamond]",
			},
		},
		{
			name: "ongoing volunteer role runs to the top and fades",
			entries: []Entry{
				role("W", "Full-time", "2025-01", "present"),
				role("V", "Volunteer", "2023-01", "present"),
			},
			want: []string{
				"W l0[above:fade node:current] l1[full:fade]",
				"V l0[] l1[above:solid node:current]",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check(t, Build(tt.entries, now, labels), tt.want)
		})
	}
}

func TestDuration(t *testing.T) {
	now := m("2026-10")
	tests := []struct {
		name, start, end, want string
	}{
		{"same month", "2024-05", "2024-05", "1 mo"},
		{"one month apart (inclusive)", "2024-05", "2024-06", "2 mos"},
		{"12 months", "2024-01", "2024-12", "1 yr"},
		{"14 months", "2023-01", "2024-02", "1 yr 2 mos"},
		{"Jul to Jul", "2023-07", "2024-07", "1 yr 1 mo"},
		{"two years", "2022-01", "2023-12", "2 yrs"},
		{"present", "2026-08", "present", "3 mos"},
	}
	for _, tt := range tests {
		g := Build([]Entry{role("X", "Full-time", tt.start, tt.end)}, now, labels)
		if got := g.Rows[0].Duration; got != tt.want {
			t.Errorf("%s: Duration = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestSummaryWithoutOverlap(t *testing.T) {
	g := Build([]Entry{role("A", "Full-time", "2020-01", "2020-12"), role("B", "Full-time", "2021-01", "2021-06")}, m("2026-10"), labels)
	if g.Summary.Text != "roles since 2020" || g.Summary.Concurrent != 1 {
		t.Errorf("summary = %+v", g.Summary)
	}
}

func TestMonth(t *testing.T) {
	v := m("2024-05")
	if v.ISO() != "2024-05" || v.Label() != "May 2024" {
		t.Errorf("Month = %s / %s", v.ISO(), v.Label())
	}
	if _, err := ParseMonth("May 2024"); err == nil {
		t.Error("want an error for a non-YYYY-MM value")
	}
}
