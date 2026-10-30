// Package career builds the view model of the experience "career graph"
// (design/experience-prototype.html) from plain entries. It is pure: "now"
// is passed in, so tests are deterministic. Templates render the result;
// nothing about the graph is hand-placed.
//
// Rows are entries sorted newest first. Lane 0 is work (Internship, Contract,
// Freelance, Full-time, Part-time) plus milestones; lane 1 is Volunteer.
// Overlapping entries in the same lane share it. Each row draws, per lane,
// line segments above/below its node (or through the row), and at most one
// merge/fork curve between lane 1 and lane 0.
package career

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Month counts months since year 0 (year*12 + month-1), so differences are
// month spans.
type Month int

// ParseMonth parses "YYYY-MM".
func ParseMonth(s string) (Month, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return 0, fmt.Errorf("career: %q is not YYYY-MM", s)
	}
	return MonthOf(t), nil
}

// MonthOf returns the month containing t.
func MonthOf(t time.Time) Month { return Month(t.Year()*12 + int(t.Month()) - 1) }

func (m Month) year() int         { return int(m) / 12 }
func (m Month) month() time.Month { return time.Month(int(m)%12 + 1) }

// ISO is the <time datetime> value: "2024-05".
func (m Month) ISO() string { return fmt.Sprintf("%04d-%02d", m.year(), int(m.month())) }

// Label is the display form: "May 2024".
func (m Month) Label() string { return m.month().String()[:3] + " " + strconv.Itoa(m.year()) }

// Kind is the row category; it is also the CSS modifier (career__row--<kind>).
type Kind string

const (
	Internship Kind = "internship"
	Volunteer  Kind = "volunteer"
	Contract   Kind = "contract"
	Freelance  Kind = "freelance"
	FullTime   Kind = "full-time"
	PartTime   Kind = "part-time"
	Milestone  Kind = "milestone"
)

// kindOrder is the legend order (the employment-type order, then milestones).
var kindOrder = []Kind{Internship, Volunteer, Contract, Freelance, FullTime, PartTime, Milestone}

// KindOf maps an employment type ("Full-time") to its Kind ("full-time").
func KindOf(typ string) Kind { return Kind(strings.ToLower(typ)) }

// lane returns 1 for volunteer roles, 0 for everything else.
func (k Kind) lane() int {
	if k == Volunteer {
		return 1
	}
	return 0
}

// Entry is one role, or a milestone (Start == End, Milestone set).
type Entry struct {
	Title, Org, Type, Icon, Location string
	Start, End                       Month // End is ignored when Present
	Present                          bool
	Bullets, Skills                  []string
	Milestone                        bool
}

// Labels is interface copy the view model needs.
type Labels struct {
	All       string // legend "All roles"
	Milestone string // type label of milestones
	Present   string // end label of an ongoing role
}

// Seg is one line segment in a row's graph cell.
type Seg struct {
	Lane   int
	Pos    string // "above" (top → node), "below" (node → bottom), "full"
	Dashed bool   // a time gap between the two engagements
	Fade   bool   // open end: the line fades out upward (still going)
}

// Class is the prototype's class list, e.g. "seg seg--l0 seg--above seg--fade".
func (s Seg) Class() string {
	c := fmt.Sprintf("seg seg--l%d seg--%s", s.Lane, s.Pos)
	if s.Dashed {
		c += " seg--dashed"
	}
	if s.Fade {
		c += " seg--fade"
	}
	return c
}

// Node is a row's marker: circle (work), diamond (volunteer), ring
// (milestone), or the larger pulsing circle of the current role.
type Node struct {
	Lane    int
	Kind    Kind
	Current bool
}

// Class is the prototype's class list, e.g. "node node--l1 node--diamond".
func (n Node) Class() string {
	c := fmt.Sprintf("node node--l%d", n.Lane)
	switch {
	case n.Current:
		c += " node--current"
	case n.Kind == Milestone:
		c += " node--ring"
	case n.Kind == Volunteer:
		c += " node--diamond"
	}
	return c
}

// Curve joins lane 1 and lane 0 in the top half of a row: a merge (lane 1
// ends into lane 0) or a fork (lane 1 leaves lane 0's node).
type Curve struct{ Fork bool }

// Path is the SVG path in the prototype's 68 × 22 viewBox (lane 0 at x=20,
// lane 1 at x=46, the node at y=22).
func (c Curve) Path() string {
	if c.Fork {
		return "M20 22 C 20 9, 46 13, 46 0"
	}
	return "M20 0 C 20 13, 46 9, 46 22"
}

// Row is one rendered entry plus its graph cell.
type Row struct {
	Entry
	Kind      Kind
	Lane      int
	Above     []Seg // rendered before the curve and node ("above", then "full")
	Curve     *Curve
	Node      Node
	Below     []Seg // rendered after the node
	TypeLabel string
	StartISO  string
	StartText string
	EndISO    string // "" when Present (or a milestone)
	EndText   string // Labels.Present when Present
	Duration  string // "1 yr 1 mo"; "" for milestones
	TenurePct int    // months / longest role's months, as a rounded %
}

// LegendItem is one legend toggle.
type LegendItem struct {
	Filter string // "all" or a Kind
	Lane   string // "all", "work", "volunteer" or "milestone": what the filter lights
	Label  string
	Count  int
}

// NodeClass is the swatch node's class list ("" for "all").
func (l LegendItem) NodeClass() string {
	switch Kind(l.Filter) {
	case Milestone:
		return "node node--ring"
	case Volunteer:
		return "node node--diamond"
	}
	if l.Filter == "all" {
		return ""
	}
	return "node"
}

// Summary is the aside's one-liner: Roles, then Text.
type Summary struct {
	Roles      int // entries that aren't milestones
	Since      int // earliest start year
	Concurrent int // most roles held in the same month
	Text       string
}

// Graph is the whole view model.
type Graph struct {
	Rows    []Row
	Legend  []LegendItem
	Summary Summary
}

type style int

const (
	none style = iota
	solid
	dashed
)

type cell struct {
	above, below, full  style
	fadeAbove, fadeFull bool
}

// Build computes the graph. now is the month "present" resolves to.
func Build(entries []Entry, now Month, l Labels) Graph {
	n := len(entries)
	rows := make([]Row, n)
	for i, e := range entries {
		k := KindOf(e.Type)
		if e.Milestone {
			k = Milestone
		}
		rows[i] = Row{Entry: e, Kind: k, Lane: k.lane(), TypeLabel: e.Type}
		if e.Milestone {
			rows[i].TypeLabel = l.Milestone
			rows[i].End = e.Start
		}
	}
	end := func(r Row) Month {
		if r.Present {
			return now
		}
		return r.End
	}
	// Newest first: by start, then by end; ties keep YAML order.
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Start != rows[j].Start {
			return rows[i].Start > rows[j].Start
		}
		return end(rows[i]) > end(rows[j])
	})

	var byLane [2][]int
	for i, r := range rows {
		byLane[r.Lane] = append(byLane[r.Lane], i)
	}
	cells := make([][2]cell, n)
	curves := make([]*Curve, n)

	// Segments between consecutive nodes of a lane: solid when continuous
	// or overlapping (the older one ran at least until the month before the
	// newer one started), dashed when there is a gap.
	for lane, idx := range byLane {
		for p := 0; p+1 < len(idx); p++ {
			a, b := idx[p], idx[p+1] // a is newer (above), b older (below)
			s := dashed
			if end(rows[b])+1 >= rows[a].Start {
				s = solid
			}
			cells[a][lane].below = s
			cells[b][lane].above = s
			for i := a + 1; i < b; i++ {
				cells[i][lane].full = s
			}
		}
	}

	// Lane 0: an ongoing newest role gets a stub that fades out upward.
	if idx := byLane[0]; len(idx) > 0 && rows[idx[0]].Present {
		cells[idx[0]][0].above, cells[idx[0]][0].fadeAbove = solid, true
	}

	// Lane 1 above its newest node: it runs up through every row that
	// started before the lane's last end, then merges into lane 0 (if that
	// row's node is in lane 0) or simply ends. An ongoing lane fades out.
	if idx := byLane[1]; len(idx) > 0 {
		t := idx[0]
		var last Month
		present := false
		for _, i := range idx {
			present = present || rows[i].Present
			if e := end(rows[i]); e > last {
				last = e
			}
		}
		k := t
		for i := t - 1; i >= 0 && (present || rows[i].Start < last); i-- {
			k = i
		}
		switch {
		case k == t && present:
			cells[t][1].above, cells[t][1].fadeAbove = solid, true
		case k < t:
			cells[t][1].above = solid
			for i := k + 1; i < t; i++ {
				cells[i][1].full = solid
			}
			switch {
			case present:
				cells[k][1].full, cells[k][1].fadeFull = solid, true
			case rows[k].Lane == 0:
				curves[k] = &Curve{}
				cells[k][1].below = solid
			default:
				cells[k][1].below = solid
			}
		}

		// Below its oldest node: fork out of the nearest lane-0 node below
		// that was active when the lane began; otherwise nothing.
		b := idx[len(idx)-1]
		for f := b + 1; f < n; f++ {
			if rows[f].Lane == 0 && end(rows[f]) >= rows[b].Start {
				cells[b][1].below = solid
				for i := b + 1; i < f; i++ {
					cells[i][1].full = solid
				}
				curves[f] = &Curve{Fork: true}
				break
			}
		}
	}

	longest := 0
	for _, r := range rows {
		if !r.Milestone {
			longest = max(longest, int(end(r)-r.Start)+1)
		}
	}
	for i := range rows {
		r := &rows[i]
		r.Node = Node{Lane: r.Lane, Kind: r.Kind, Current: r.Present}
		r.Curve = curves[i]
		for lane, c := range cells[i] {
			if c.above != none {
				r.Above = append(r.Above, Seg{Lane: lane, Pos: "above", Dashed: c.above == dashed, Fade: c.fadeAbove})
			}
		}
		for lane, c := range cells[i] {
			if c.full != none {
				r.Above = append(r.Above, Seg{Lane: lane, Pos: "full", Dashed: c.full == dashed, Fade: c.fadeFull})
			}
		}
		for lane, c := range cells[i] {
			if c.below != none {
				r.Below = append(r.Below, Seg{Lane: lane, Pos: "below", Dashed: c.below == dashed})
			}
		}
		r.StartISO, r.StartText = r.Start.ISO(), r.Start.Label()
		switch {
		case r.Milestone:
		case r.Present:
			r.EndText = l.Present
		default:
			r.EndISO, r.EndText = r.End.ISO(), r.End.Label()
		}
		if !r.Milestone {
			months := int(end(*r)-r.Start) + 1
			r.Duration = Duration(months)
			r.TenurePct = int(math.Round(float64(months) / float64(longest) * 100))
		}
	}

	return Graph{Rows: rows, Legend: legend(rows, l), Summary: summary(rows, now)}
}

// Duration formats an inclusive month count: "9 mos", "1 yr 1 mo", "2 yrs".
func Duration(months int) string {
	months = max(months, 1)
	y, m := months/12, months%12
	var parts []string
	if y > 0 {
		parts = append(parts, plural(y, "yr"))
	}
	if m > 0 {
		parts = append(parts, plural(m, "mo"))
	}
	return strings.Join(parts, " ")
}

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return strconv.Itoa(n) + " " + unit + "s"
}

func legend(rows []Row, l Labels) []LegendItem {
	counts := map[Kind]int{}
	for _, r := range rows {
		counts[r.Kind]++
	}
	items := []LegendItem{{Filter: "all", Lane: "all", Label: l.All, Count: len(rows)}}
	for _, k := range kindOrder {
		if counts[k] == 0 {
			continue
		}
		item := LegendItem{Filter: string(k), Lane: "work", Count: counts[k]}
		switch k {
		case Volunteer:
			item.Lane = "volunteer"
		case Milestone:
			item.Lane = "milestone"
		}
		for _, r := range rows {
			if r.Kind == k {
				item.Label = r.TypeLabel
				break
			}
		}
		items = append(items, item)
	}
	return items
}

var smallNumbers = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

func summary(rows []Row, now Month) Summary {
	var s Summary
	var first, last Month
	for _, r := range rows {
		if r.Milestone {
			continue
		}
		e := r.End
		if r.Present {
			e = now
		}
		if s.Roles == 0 || r.Start < first {
			first = r.Start
		}
		if s.Roles == 0 || e > last {
			last = e
		}
		s.Roles++
	}
	if s.Roles == 0 {
		return s
	}
	s.Since = first.year()
	for m := first; m <= last; m++ {
		active := 0
		for _, r := range rows {
			e := r.End
			if r.Present {
				e = now
			}
			if !r.Milestone && r.Start <= m && m <= e {
				active++
			}
		}
		s.Concurrent = max(s.Concurrent, active)
	}
	noun := "roles"
	if s.Roles == 1 {
		noun = "role"
	}
	s.Text = fmt.Sprintf("%s since %d", noun, s.Since)
	if s.Concurrent > 1 {
		word := strconv.Itoa(s.Concurrent)
		if s.Concurrent < len(smallNumbers) {
			word = smallNumbers[s.Concurrent]
		}
		s.Text += fmt.Sprintf(", up to %s at once", word)
	}
	return s
}
