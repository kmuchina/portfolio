// Package evidence finds where each skill was used: in which projects (by
// their stack) and in which jobs (by their skills). It is pure; templates
// render the result as skill buttons with a "used in" panel.
//
// Matching is exact after normalising both sides: lower-case, and every
// character other than a letter, a digit, "+" or "#" dropped, so "Node.js"
// matches "nodejs" and "Git & GitHub" matches "git github", while "C++"
// stays distinct from "C". Nothing fuzzier than that. A place counts once
// per skill, however many of its entries match.
package evidence

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kmuchina/portfolio/internal/site"
)

// Item is one place a skill was used.
type Item struct {
	Label     string // "Apple Care", "ICT Officer, Kenya Roads Board"
	Href      string // "#project-apple-care", "#job-kenya-roads-board"
	Kind      string // "Project", or the job's employment type
	Project   bool
	Volunteer bool
}

// Skill is one pill: its evidence, the panel id and the panel's label.
type Skill struct {
	site.SkillItem
	PanelID  string
	Evidence []Item
	Label    string // "Kotlin, used in"; "" without evidence
	Count    string // "used in 2 places": the button's visually hidden count
}

// Group is a skill group with its skills resolved.
type Group struct {
	site.SkillGroup
	Skills []Skill
}

// Labels words the panel label; see Verb.
type Labels struct {
	UsedIn, UsedAt, ShownAt, Project string
	Place, Places                    string // "place" / "places" (Skill.Count)
}

// Normalize is the matching key of a skill name.
func Normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' || r == '#' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Build resolves every skill group of s.
func Build(s *site.Site, l Labels) []Group {
	type place struct {
		item  Item
		terms map[string]bool
	}
	var places []place
	terms := func(list []string) map[string]bool {
		m := map[string]bool{}
		for _, t := range list {
			if k := Normalize(t); k != "" {
				m[k] = true
			}
		}
		return m
	}
	for _, p := range s.Projects.Items {
		places = append(places, place{
			item:  Item{Label: p.Title, Href: "#project-" + p.Slug, Kind: l.Project, Project: true},
			terms: terms(p.Stack),
		})
	}
	for _, j := range s.Experience.Items {
		org, _, _ := strings.Cut(j.Org, ",")
		places = append(places, place{
			item:  Item{Label: j.Title + ", " + strings.TrimSpace(org), Href: "#" + j.ID(), Kind: j.Type, Volunteer: j.Type == "Volunteer"},
			terms: terms(j.Skills),
		})
	}

	groups := make([]Group, 0, len(s.Skills.Groups))
	for _, g := range s.Skills.Groups {
		out := Group{SkillGroup: g}
		for _, it := range g.Items {
			sk := Skill{SkillItem: it, PanelID: g.PanelID(it)}
			key := Normalize(it.Name)
			for _, p := range places {
				if key != "" && p.terms[key] {
					sk.Evidence = append(sk.Evidence, p.item)
				}
			}
			if len(sk.Evidence) > 0 {
				verb, n := Verb(sk.Evidence, l), len(sk.Evidence)
				sk.Label = it.Name + ", " + verb
				noun := l.Places
				if n == 1 {
					noun = l.Place
				}
				sk.Count = fmt.Sprintf("%s %d %s", verb, n, noun)
			}
			out.Skills = append(out.Skills, sk)
		}
		groups = append(groups, out)
	}
	return groups
}

// Verb picks the panel's wording: "used in" when every place is a project,
// "used at" when every place is a paid role (not volunteer), and "shown at"
// otherwise: a mix, or any volunteer role (leadership shown, not a tool used).
func Verb(ev []Item, l Labels) string {
	projects, paid := 0, 0
	for _, e := range ev {
		switch {
		case e.Project:
			projects++
		case !e.Volunteer:
			paid++
		}
	}
	switch len(ev) {
	case projects:
		return l.UsedIn
	case paid:
		return l.UsedAt
	}
	return l.ShownAt
}
