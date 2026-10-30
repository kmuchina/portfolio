// Package site is the content model: structs mirroring content/site.yaml,
// strict loading and validation.
package site

import (
	"html/template"
	"strings"
	"time"
)

// Site is the root of content/site.yaml.
type Site struct {
	Meta       Meta       `yaml:"site"`
	Person     Person     `yaml:"person"`
	Contact    Contact    `yaml:"contact"`
	About      About      `yaml:"about"`
	Experience Experience `yaml:"experience"`
	Skills     Skills     `yaml:"skills"`
	Projects   Projects   `yaml:"projects"`
	Education  Education  `yaml:"education"`
	Languages  []Language `yaml:"languages"`
	UI         UI         `yaml:"ui"`
}

// Meta is the `site:` block.
type Meta struct {
	URL         string `yaml:"url"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Lang        string `yaml:"lang"`
	ShowPhone   bool   `yaml:"show_phone"`
}

type Person struct {
	Greeting     string `yaml:"greeting"`
	FirstName    string `yaml:"first_name"`
	MiddleName   string `yaml:"middle_name"`
	LastName     string `yaml:"last_name"`
	Role         string `yaml:"role"`
	Location     string `yaml:"location"`
	Availability string `yaml:"availability"`
	Quote        string `yaml:"quote"`
	Summary      string `yaml:"summary"`
	Photo        string `yaml:"photo"`
	PhotoAlt     string `yaml:"photo_alt"`
	ResumePDF    string `yaml:"resume_pdf"`
}

// FullName returns "First Middle Last", skipping empty parts.
func (p Person) FullName() string {
	return strings.Join(strings.Fields(p.FirstName+" "+p.MiddleName+" "+p.LastName), " ")
}

type Contact struct {
	Eyebrow string `yaml:"eyebrow"`
	Heading string `yaml:"heading"`
	Body    string `yaml:"body"`
	Email   string `yaml:"email"`
	Phone   string `yaml:"phone"`
	Links   []Link `yaml:"links"`
}

// TelHref turns "+254 700 000 000" into "tel:+254700000000". It returns
// template.URL because html/template rejects the tel: scheme; the value is
// built from digits and a leading "+" only, so it is safe.
func (c Contact) TelHref() template.URL {
	var b strings.Builder
	for i, r := range strings.TrimSpace(c.Phone) {
		if (r >= '0' && r <= '9') || (r == '+' && i == 0) {
			b.WriteRune(r)
		}
	}
	return template.URL("tel:" + b.String())
}

// MailtoHref returns "mailto:<email>".
func (c Contact) MailtoHref() string { return "mailto:" + c.Email }

// Link is an external link (profiles, project links).
type Link struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
	Icon  string `yaml:"icon"`
}

// Action is an optional section-header link ("View full résumé →"). URL is
// either an absolute https:// URL or a file path under static/ (rendered
// through the asset manifest).
type Action struct {
	Label string `yaml:"label"`
	URL   string `yaml:"url"`
}

// IsExternal reports whether URL is a full URL rather than a static file.
func (a Action) IsExternal() bool { return strings.Contains(a.URL, "://") }

// AssetPath is URL as an asset-manifest key: a static file path without a
// leading slash ("/Kennedy_Muchina_Resume.pdf" → "Kennedy_Muchina_Resume.pdf").
func (a Action) AssetPath() string { return strings.TrimPrefix(a.URL, "/") }

type About struct {
	Eyebrow string      `yaml:"eyebrow"`
	Heading string      `yaml:"heading"`
	Body    string      `yaml:"body"`
	Focus   []FocusItem `yaml:"focus"`
	Facts   []Fact      `yaml:"facts"`
}

type FocusItem struct {
	Label string `yaml:"label"`
	Icon  string `yaml:"icon"`
}

type Fact struct {
	Title  string `yaml:"title"`
	Detail string `yaml:"detail"`
	Icon   string `yaml:"icon"`
}

type Experience struct {
	Eyebrow string  `yaml:"eyebrow"`
	Heading string  `yaml:"heading"`
	Action  *Action `yaml:"action"`
	Items   []Job   `yaml:"items"`
}

// Job is one experience entry. Dates are "YYYY-MM"; End may be "present".
// Type is one of EmploymentTypes; "intern" never appears in Title (the
// employment type says it). Bullets may use **bold** (internal/textfmt).
type Job struct {
	Title    string   `yaml:"title"`
	Org      string   `yaml:"org"`
	Type     string   `yaml:"type"`
	Icon     string   `yaml:"icon"`
	Location string   `yaml:"location"`
	Start    string   `yaml:"start"`
	End      string   `yaml:"end"`
	Bullets  []string `yaml:"bullets"`
	Skills   []string `yaml:"skills"`
}

// EmploymentTypes are the allowed experience `type` values (as on LinkedIn).
var EmploymentTypes = []string{"Internship", "Volunteer", "Contract", "Freelance", "Full-time", "Part-time"}

// Present is the `end` value of an ongoing role.
const Present = "present"

type Skills struct {
	Eyebrow string       `yaml:"eyebrow"`
	Heading string       `yaml:"heading"`
	Groups  []SkillGroup `yaml:"groups"`
}

type SkillGroup struct {
	Name  string   `yaml:"name"`
	Icon  string   `yaml:"icon"`
	Items []string `yaml:"items"`
}

type Projects struct {
	Eyebrow string    `yaml:"eyebrow"`
	Heading string    `yaml:"heading"`
	Action  *Action   `yaml:"action"`
	Items   []Project `yaml:"items"`
}

type Project struct {
	Slug     string   `yaml:"slug"`
	Title    string   `yaml:"title"`
	Kind     string   `yaml:"kind"`
	Stack    []string `yaml:"stack"`
	Summary  string   `yaml:"summary"`
	Metric   string   `yaml:"metric"`
	Image    string   `yaml:"image"`
	ImageAlt string   `yaml:"image_alt"`
	Links    []Link   `yaml:"links"`
}

// HasLinks reports whether the project has any links to render.
func (p Project) HasLinks() bool { return len(p.Links) > 0 }

type Education struct {
	Eyebrow string   `yaml:"eyebrow"`
	Heading string   `yaml:"heading"`
	Items   []School `yaml:"items"`
}

type School struct {
	Degree     string     `yaml:"degree"`
	School     string     `yaml:"school"`
	Date       string     `yaml:"date"`
	Highlights []string   `yaml:"highlights"`
	Timeline   *Milestone `yaml:"timeline"` // optional: also a milestone in the career graph
}

// Milestone is a dated point in the career graph (e.g. graduation).
type Milestone struct {
	Date   string `yaml:"date"` // "YYYY-MM"
	Title  string `yaml:"title"`
	Detail string `yaml:"detail"`
}

// DateISO returns the date as a <time datetime> value ("2025-12"), or "" when
// it isn't a single month (e.g. "2016 – 2019").
func (s School) DateISO() string { return monthISO(s.Date) }

type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
}

// UI holds interface copy (labels, nav, buttons) so templates never
// hard-code text.
type UI struct {
	SkipLink         string    `yaml:"skip_link"`
	NavLabel         string    `yaml:"nav_label"`
	HomeLabel        string    `yaml:"home_label"`
	Nav              []NavItem `yaml:"nav"`
	Menu             MenuUI    `yaml:"menu"`
	Theme            ThemeUI   `yaml:"theme"`
	CTAProjects      string    `yaml:"cta_projects"`
	CTAResume        string    `yaml:"cta_resume"`
	Current          string    `yaml:"current"`
	LanguagesHeading string    `yaml:"languages_heading"`
	SocialLabel      string    `yaml:"social_label"`
	Career           CareerUI  `yaml:"career"`
	NotFound         NotFound  `yaml:"not_found"`
}

type NavItem struct {
	Label string `yaml:"label"`
	Href  string `yaml:"href"`
}

// CareerUI is the interface copy of the experience graph's aside and entries.
type CareerUI struct {
	Highlight   string `yaml:"highlight"`    // legend heading ("Highlight")
	LegendLabel string `yaml:"legend_label"` // legend group's accessible name
	All         string `yaml:"all"`          // "All roles"
	Milestone   string `yaml:"milestone"`    // type label of milestones
	Present     string `yaml:"present"`      // end label of an ongoing role
	SkillsLabel string `yaml:"skills_label"` // accessible name of each tag list
}

// MenuUI labels the mobile menu dialog and its open/close buttons.
type MenuUI struct {
	Label string `yaml:"label"`
	Open  string `yaml:"open"`
	Close string `yaml:"close"`
}

// ThemeUI labels the theme toggle. Each instance on a page needs its own
// radiogroup label: Label in the header, MenuLabel in the mobile menu.
type ThemeUI struct {
	Label     string `yaml:"label"`
	MenuLabel string `yaml:"menu_label"`
	Dark      string `yaml:"dark"`
	System    string `yaml:"system"`
	Light     string `yaml:"light"`
}

type NotFound struct {
	Title string `yaml:"title"`
	Body  string `yaml:"body"`
	Back  string `yaml:"back"`
}

// monthISO parses "Aug 2026" into "2026-08"; anything else returns "".
func monthISO(s string) string {
	t, err := time.Parse("Jan 2006", strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	return t.Format("2006-01")
}
