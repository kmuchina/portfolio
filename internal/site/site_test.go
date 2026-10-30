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

func TestTelHref(t *testing.T) {
	tests := []struct{ in, want string }{
		{"+254 700 000 000", "tel:+254700000000"},
		{"0700-000-000", "tel:0700000000"},
		{" +1 (555) 010 9999 ", "tel:+15550109999"},
		{"", "tel:"},
	}
	for _, tt := range tests {
		if got := string((Contact{Phone: tt.in}).TelHref()); got != tt.want {
			t.Errorf("TelHref(%q) = %q, want %q", tt.in, got, tt.want)
		}
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
