package evidence

import (
	"testing"

	"github.com/kmuchina/portfolio/internal/site"
)

var labels = Labels{UsedIn: "used in", UsedAt: "used at", ShownAt: "shown at", Project: "Project", Place: "place", Places: "places"}

// The real content: structural invariants only, so copy edits don't break it.
func TestRealEvidence(t *testing.T) {
	s, err := site.Load("../../content/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, p := range s.Projects.Items {
		ids["#project-"+p.Slug] = true
	}
	for _, j := range s.Experience.Items {
		ids["#"+j.ID()] = true
	}
	total := 0
	for _, g := range Build(s, labels) {
		for _, sk := range g.Skills {
			if (len(sk.Evidence) == 0) != (sk.Label == "" && sk.Count == "") {
				t.Errorf("%s: label %q / count %q inconsistent with %d places", sk.Name, sk.Label, sk.Count, len(sk.Evidence))
			}
			for _, e := range sk.Evidence {
				total++
				if !ids[e.Href] {
					t.Errorf("%s: evidence link %q has no target", sk.Name, e.Href)
				}
			}
		}
	}
	if total == 0 {
		t.Error("no skill has any evidence")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct{ a, b string }{
		{"Node.js", "nodejs"},
		{"Git & GitHub", "git github"},
		{"  Jetpack   Compose ", "jetpack-compose"},
		{"REST APIs", "rest-apis"},
		{"TensorFlow Lite", "tensorflowlite"},
	}
	for _, tt := range tests {
		if Normalize(tt.a) != Normalize(tt.b) {
			t.Errorf("%q and %q should match (%q vs %q)", tt.a, tt.b, Normalize(tt.a), Normalize(tt.b))
		}
	}
	for _, tt := range []struct{ a, b string }{{"C++", "C"}, {"C#", "C"}, {"Java", "JavaScript"}} {
		if Normalize(tt.a) == Normalize(tt.b) {
			t.Errorf("%q and %q must not match", tt.a, tt.b)
		}
	}
}

func synthetic() *site.Site {
	return &site.Site{
		Skills: site.Skills{Groups: []site.SkillGroup{{
			Name: "Tools",
			Items: []site.SkillItem{
				{Name: "Node.js"}, {Name: "Git & GitHub"}, {Name: "Mentoring"}, {Name: "Go"}, {Name: "Cobol"},
			},
		}}},
		Projects: site.Projects{Items: []site.Project{
			// Listed twice, in two spellings: still one place.
			{Slug: "api", Title: "API", Stack: []string{"node.js", "NodeJS", "Go"}},
		}},
		Experience: site.Experience{Items: []site.Job{
			{Title: "Engineer", Org: "Acme, Inc.", Type: "Full-time", Skills: []string{"git github", "Go"}},
			{Title: "Mentor", Org: "Club", Type: "Volunteer", Skills: []string{"Mentoring", "Go"}},
		}},
	}
}

func TestSynthetic(t *testing.T) {
	got := map[string]Skill{}
	for _, sk := range Build(synthetic(), labels)[0].Skills {
		got[sk.Name] = sk
	}
	tests := []struct {
		skill string
		n     int
		label string
	}{
		{"Node.js", 1, "Node.js, used in"},           // duplicate match counted once
		{"Git & GitHub", 1, "Git & GitHub, used at"}, // paid role only
		{"Mentoring", 1, "Mentoring, shown at"},      // volunteer role
		{"Go", 3, "Go, shown at"},                    // mixed
		{"Cobol", 0, ""},                             // no evidence
	}
	for _, tt := range tests {
		sk := got[tt.skill]
		if len(sk.Evidence) != tt.n || sk.Label != tt.label {
			t.Errorf("%s: got %d %q, want %d %q", tt.skill, len(sk.Evidence), sk.Label, tt.n, tt.label)
		}
	}
	if href := got["Git & GitHub"].Evidence[0].Href; href != "#job-acme" {
		t.Errorf("job href = %q, want #job-acme", href)
	}
	if id := got["Git & GitHub"].PanelID; id != "used-tools-git-github" {
		t.Errorf("panel id = %q", id)
	}
}

func TestEmpty(t *testing.T) {
	if g := Build(&site.Site{}, labels); len(g) != 0 {
		t.Errorf("no groups: got %v", g)
	}
	s := synthetic()
	s.Projects.Items, s.Experience.Items = nil, nil
	for _, sk := range Build(s, labels)[0].Skills {
		if len(sk.Evidence) != 0 || sk.Label != "" {
			t.Errorf("%s: want no evidence without projects or jobs", sk.Name)
		}
	}
}
