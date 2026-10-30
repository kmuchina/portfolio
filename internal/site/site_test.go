package site

import (
	"strings"
	"testing"
)

const (
	contentPath = "../../content/site.yaml"
	staticDir   = "../../static"
	iconDir     = "../../templates/icons"
)

func TestRealContentValidates(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(s, staticDir, iconDir); err != nil {
		t.Fatal(err)
	}
	if got := s.Person.FullName(); got != "Kennedy Muchina Kinuthia" {
		t.Errorf("FullName = %q", got)
	}
	if s.Contact.Email != "muchina537@gmail.com" {
		t.Errorf("email = %q", s.Contact.Email)
	}
}

func TestLanguageMeter(t *testing.T) {
	tests := []struct {
		level string
		meter int
		label string
	}{
		{"Elementary", 1, "Level 1 of 5"},
		{"Limited working", 2, "Level 2 of 5"},
		{"Professional working", 3, "Level 3 of 5"},
		{"Full professional", 4, "Level 4 of 5"},
		{"Native", 5, "Level 5 of 5"},
		{"Fluent", 0, "Level 0 of 5"},
	}
	for _, tt := range tests {
		l := Language{Name: "X", Level: tt.level}
		if l.Meter() != tt.meter || l.MeterLabel() != tt.label {
			t.Errorf("%s: meter %d %q, want %d %q", tt.level, l.Meter(), l.MeterLabel(), tt.meter, tt.label)
		}
		lit := 0
		for _, on := range l.MeterBars() {
			if on {
				lit++
			}
		}
		if len(l.MeterBars()) != 5 || lit != tt.meter {
			t.Errorf("%s: bars %v", tt.level, l.MeterBars())
		}
	}
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Languages[0].Meter(); got != 3 { // English, Professional working
		t.Errorf("English meter = %d, want 3", got)
	}
	s.Languages[0].Level = "Professional working proficiency"
	if _, err := Validate(s, staticDir, iconDir); err == nil || !strings.Contains(err.Error(), "must be one of Elementary") {
		t.Fatalf("want a language level error, got %v", err)
	}
}

func TestFactsCurrently(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range s.Facts() {
		got = append(got, f.Label+": "+f.Value+" / "+f.Detail)
	}
	want := "Currently: ICT & Digital Innovation / Kenya Flower Council | " +
		"Education: BSc Computer Science, First Class / Moi University, 2025 | " +
		"Based in: Nairobi, Kenya / Open to remote & global roles | " +
		"Leadership: IEEE branch chair, project lead / Team player, mentor & community builder"
	if strings.Join(got, " | ") != want {
		t.Errorf("facts:\n got  %s\n want %s", strings.Join(got, " | "), want)
	}
	// No current role: no "Currently" fact.
	s.Experience.Items[0].End = "2026-09"
	if f := s.Facts(); len(f) != 3 || f[0].Label != "Education" {
		t.Errorf("without a current role, want the 3 listed facts, got %v", f)
	}
}

func TestJobIDs(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, j := range s.Experience.Items {
		got = append(got, j.ID())
	}
	want := "job-kenya-flower-council job-kenya-roads-board job-ieee-tech4good job-ieee-moi-university-student-branch"
	if strings.Join(got, " ") != want {
		t.Errorf("job ids = %v", got)
	}
	s.Experience.Items[1].Org = "Kenya Flower Council, Karen"
	if _, err := Validate(s, staticDir, iconDir); err == nil || !strings.Contains(err.Error(), "is not unique") {
		t.Fatalf("want a duplicate job id error, got %v", err)
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Git & GitHub":     "git-github",
		"IEEE Tech4Good":   "ieee-tech4good",
		"  Node.js  ":      "node-js",
		"Machine learning": "machine-learning",
		"C++":              "c",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSectionRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(s *Site)
		want   string
	}{
		{"skill icon missing", func(s *Site) { s.Skills.Groups[0].Items[0].Icon = "no-such-icon" }, `icon "no-such-icon"`},
		{"skill icon empty", func(s *Site) { s.Skills.Groups[0].Items[0].Icon = "" }, "every skill pill has an icon"},
		{"duplicate skill", func(s *Site) { s.Skills.Groups[0].Items[1].Name = "kotlin" }, "panel id"},
		{"bad placeholder", func(s *Site) { s.Projects.Items[0].Placeholder = "phone" }, "must be one of journal | scanner | none"},
		{"bad time zone", func(s *Site) { s.Contact.Timezone.IANA = "Africa/Atlantis" }, "is not a time zone"},
		{"source url", func(s *Site) { s.Meta.SourceURL = "http://github.com/x" }, "site.source_url"},
		{"education date", func(s *Site) { s.Education.Items[0].Date = "Dec 2025" }, "must be YYYY-MM"},
		{"education years", func(s *Site) { s.Education.Items[1].End = "19" }, "must both be YYYY"},
		{"education neither", func(s *Site) { s.Education.Items[1].Grades = nil }, "needs honours"},
		{"metric half", func(s *Site) { s.Projects.Items[0].Metric.Label = "" }, "metric: value and label"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Load(contentPath)
			if err != nil {
				t.Fatal(err)
			}
			tt.mutate(s)
			if _, err := Validate(s, staticDir, iconDir); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestEducationViews(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	if d := s.Education.Degree(); d == nil || d.School != "Moi University" || d.DateText() != "Dec 2025" {
		t.Errorf("degree = %+v", d)
	}
	if sc := s.Education.Schools(); len(sc) != 1 || sc[0].School != "Njiiri High School" {
		t.Errorf("schools = %+v", sc)
	}
}

func TestMissingScreenshotWarns(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Projects.Items[0].Screenshot = "img/projects/does-not-exist.png"
	warnings, err := Validate(s, staticDir, iconDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) == 0 || !strings.Contains(strings.Join(warnings, "\n"), "does-not-exist.png") {
		t.Errorf("want a screenshot warning, got %v", warnings)
	}
}

func TestExperienceRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(j *Job)
		want   string // substring of the error; "" = valid
	}{
		{"real entry is valid", func(j *Job) {}, ""},
		{"intern in title", func(j *Job) { j.Title = "ICT Intern" }, `must not contain "intern"`},
		{"internship in title", func(j *Job) { j.Title = "Internship Programme" }, `must not contain "intern"`},
		{"case-insensitive", func(j *Job) { j.Title = "INTERN" }, `must not contain "intern"`},
		{"International is fine", func(j *Job) { j.Title = "International Liaison" }, ""},
		{"unknown type", func(j *Job) { j.Type = "Apprenticeship" }, "must be one of Internship | Volunteer"},
		{"lower-case type", func(j *Job) { j.Type = "internship" }, "must be one of"},
		{"bad start", func(j *Job) { j.Start = "Aug 2026" }, "must be YYYY-MM"},
		{"bad month", func(j *Job) { j.Start = "2026-13" }, "must be YYYY-MM"},
		{"end before start", func(j *Job) { j.Start, j.End = "2025-01", "2024-05" }, "is before start"},
		{"present end", func(j *Job) { j.End = Present }, ""},
		{"missing icon file", func(j *Job) { j.Icon = "no-such-icon" }, `icon "no-such-icon"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Load(contentPath)
			if err != nil {
				t.Fatal(err)
			}
			tt.mutate(&s.Experience.Items[1])
			_, err = Validate(s, staticDir, iconDir)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("want valid, got %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("want error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestNoInternInRealTitles(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, j := range s.Experience.Items {
		if internRe.MatchString(j.Title) {
			t.Errorf("title %q contains intern", j.Title)
		}
	}
}

func TestMilestoneDate(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Education.Items[0].Timeline.Date = "Dec 2025"
	if _, err := Validate(s, staticDir, iconDir); err == nil || !strings.Contains(err.Error(), "timeline.date") {
		t.Fatalf("want a timeline.date error, got %v", err)
	}
}

func TestActionAssetPath(t *testing.T) {
	if got := (Action{URL: "/Kennedy_Muchina_Resume.pdf"}).AssetPath(); got != "Kennedy_Muchina_Resume.pdf" {
		t.Errorf("AssetPath = %q", got)
	}
}

func TestUnknownFieldFails(t *testing.T) {
	_, err := Parse(strings.NewReader("person:\n  first_nmae: Kennedy\n"))
	if err == nil || !strings.Contains(err.Error(), "first_nmae") {
		t.Fatalf("want unknown-field error, got %v", err)
	}
}

func TestRetiredEmailFails(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.About.Body += " Write to kmuchina22@gmail.com."
	_, err = Validate(s, staticDir, iconDir)
	if err == nil || !strings.Contains(err.Error(), "kmuchina22") {
		t.Fatalf("want retired-email error, got %v", err)
	}
}

func TestValidateCollectsAllProblems(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Contact.Email = "not-an-email"
	s.Contact.Links[0].URL = "http://insecure.example"
	s.Projects.Items[1].Slug = s.Projects.Items[0].Slug
	s.About.Heading = ""
	_, err = Validate(s, staticDir, iconDir)
	probs, ok := err.(Problems)
	if !ok || len(probs) != 4 {
		t.Fatalf("want 4 problems, got %v", err)
	}
}

func TestActionValidation(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Experience.Action = &Action{Label: "Résumé", URL: "missing.pdf"}
	s.Projects.Action = &Action{URL: "http://insecure.example"}
	_, err = Validate(s, staticDir, iconDir)
	probs, ok := err.(Problems)
	if !ok || len(probs) != 3 { // missing file, missing label, non-https URL
		t.Fatalf("want 3 problems, got %v", err)
	}
}

func TestResumeOptional(t *testing.T) {
	s, err := Load(contentPath)
	if err != nil {
		t.Fatal(err)
	}
	s.Person.ResumePDF = ""
	if _, err := Validate(s, staticDir, iconDir); err != nil {
		t.Errorf("empty resume_pdf should be valid, got %v", err)
	}
	s.Person.ResumePDF = "missing.pdf"
	if _, err := Validate(s, staticDir, iconDir); err == nil {
		t.Error("a resume_pdf that doesn't exist under static/ should fail")
	}
}

func TestActionIsExternal(t *testing.T) {
	if !(Action{URL: "https://github.com/kmuchina"}).IsExternal() {
		t.Error("https URL should be external")
	}
	if (Action{URL: "Kennedy_Muchina_Resume.pdf"}).IsExternal() {
		t.Error("static path should not be external")
	}
}
