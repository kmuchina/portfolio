package evidence

import (
	"strings"
	"testing"

	"github.com/kmuchina/portfolio/internal/site"
)

var labels = Labels{UsedIn: "used in", UsedAt: "used at", ShownAt: "shown at", Project: "Project"}

// The real content: which skills have evidence, how much, and where.
func TestRealEvidence(t *testing.T) {
	s, err := site.Load("../../content/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct {
		count int
		where string
		label string
	}{
		"Kotlin":           {1, "Prayer & Gratitude Journal", "Kotlin, used in"},
		"Jetpack Compose":  {1, "Prayer & Gratitude Journal", "Jetpack Compose, used in"},
		"Python":           {1, "Apple Care", "Python, used in"},
		"TensorFlow Lite":  {1, "Apple Care", "TensorFlow Lite, used in"},
		"CNN":              {1, "Apple Care", "CNN, used in"},
		"Business Central": {1, "ICT Officer, Kenya Roads Board", "Business Central, used at"},
		"Cisco networking": {1, "ICT Officer, Kenya Roads Board", "Cisco networking, used at"},
		"Team leadership":  {2, "Project Lead & Tech Lead, IEEE Tech4Good; Chairperson, IEEE Moi University Student Branch", "Team leadership, shown at"},
	}
	seen := 0
	for _, g := range Build(s, labels) {
		for _, sk := range g.Skills {
			w, ok := want[sk.Name]
			if !ok {
				if len(sk.Evidence) != 0 {
					t.Errorf("%s: want no evidence, got %v", sk.Name, sk.Evidence)
				}
				continue
			}
			seen++
			var where []string
			for _, e := range sk.Evidence {
				where = append(where, e.Label)
			}
			if len(sk.Evidence) != w.count || strings.Join(where, "; ") != w.where || sk.Label != w.label {
				t.Errorf("%s: got %d %q %q, want %d %q %q", sk.Name, len(sk.Evidence), strings.Join(where, "; "), sk.Label, w.count, w.where, w.label)
			}
			t.Logf("%-18s %d  %s", sk.Name, len(sk.Evidence), strings.Join(where, "; "))
		}
	}
	if seen != len(want) {
		t.Errorf("found %d of the %d expected skills", seen, len(want))
	}
}

func TestEvidenceHrefs(t *testing.T) {
	s, err := site.Load("../../content/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	hrefs := map[string]string{}
	for _, g := range Build(s, labels) {
		for _, sk := range g.Skills {
			for _, e := range sk.Evidence {
				hrefs[e.Label] = e.Href + " " + e.Kind
			}
		}
	}
	for label, want := range map[string]string{
		"Apple Care":                                      "#project-apple-care Project",
		"ICT Officer, Kenya Roads Board":                  "#job-kenya-roads-board Internship",
		"Chairperson, IEEE Moi University Student Branch": "#job-ieee-moi-university-student-branch Volunteer",
	} {
		if hrefs[label] != want {
			t.Errorf("%s: got %q, want %q", label, hrefs[label], want)
		}
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
