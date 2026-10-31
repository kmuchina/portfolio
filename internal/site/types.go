// Package site is the content model: structs mirroring content/site.yaml,
// strict loading and validation.
package site

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Site is the root of content/site.yaml.
type Site struct {
	Meta       Meta       `yaml:"site"`
	SEO        SEO        `yaml:"seo"`
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
}

// SEO is the `seo:` block: what search engines and link previews read
// (Open Graph, the Person JSON-LD, the web manifest).
type SEO struct {
	Locale     string   `yaml:"locale"`      // og:locale: "en_KE"
	ShortName  string   `yaml:"short_name"`  // web manifest short_name: "Kennedy M."
	Image      string   `yaml:"image"`       // share image under static/: a 1200 × 630 PNG
	ImageAlt   string   `yaml:"image_alt"`   // og:image:alt
	Address    Address  `yaml:"address"`     // JSON-LD PostalAddress
	KnowsAbout []string `yaml:"knows_about"` // skill groups whose items become JSON-LD knowsAbout
}

// Address is a city and an ISO 3166-1 alpha-2 country code.
type Address struct {
	Locality string `yaml:"locality"` // "Nairobi"
	Country  string `yaml:"country"`  // "KE"
}

// Share image size (Open Graph's recommended 1.91:1).
const (
	ShareImageW = 1200
	ShareImageH = 630
)

// Description length search engines show in full (validation warns outside it).
const (
	DescriptionMin = 150
	DescriptionMax = 160
)

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
	Eyebrow  string   `yaml:"eyebrow"`
	Heading  string   `yaml:"heading"`
	Body     string   `yaml:"body"`
	Email    string   `yaml:"email"`
	Timezone Timezone `yaml:"timezone"`
	Links    []Link   `yaml:"links"`
}

// Timezone drives the contact section's live local time: IANA is a
// time.LoadLocation name ("Africa/Nairobi"), Label its display ("EAT, UTC+3").
type Timezone struct {
	IANA  string `yaml:"iana"`
	Label string `yaml:"label"`
}

// MailtoHref returns "mailto:<email>".
func (c Contact) MailtoHref() string { return "mailto:" + c.Email }

// Link is an external link (profiles, project links).
type Link struct {
	Label  string `yaml:"label"`
	Handle string `yaml:"handle"` // contact channels: "in/kmuchina"
	URL    string `yaml:"url"`
	Icon   string `yaml:"icon"`
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

// IsPDF reports whether the action links to a PDF (it gets a hidden "(PDF)").
func (a Action) IsPDF() bool { return strings.HasSuffix(strings.ToLower(a.URL), ".pdf") }

type About struct {
	Eyebrow string      `yaml:"eyebrow"` // optional: none is rendered when empty
	Heading string      `yaml:"heading"`
	Body    string      `yaml:"body"` // **bold** allowed (internal/textfmt)
	Focus   []FocusItem `yaml:"focus"`
	Facts   []Fact      `yaml:"facts"` // "Currently" is derived (Site.Facts), not listed
}

type FocusItem struct {
	Label  string `yaml:"label"`
	Icon   string `yaml:"icon"`
	Detail string `yaml:"detail"`
}

// Fact is one row of the about fact sheet: "Based in" → "Nairobi, Kenya",
// with a quieter detail line under the value.
type Fact struct {
	Label  string `yaml:"label"`
	Value  string `yaml:"value"`
	Detail string `yaml:"detail"`
}

// Facts is the about fact sheet: "Currently" (the job whose end is
// "present": its title, then its org) first, when there is one, then
// about.facts in order.
func (s *Site) Facts() []Fact {
	var out []Fact
	for _, j := range s.Experience.Items {
		if j.End == Present {
			out = append(out, Fact{Label: s.UI.About.Currently, Value: j.Title, Detail: j.Org})
			break
		}
	}
	return append(out, s.About.Facts...)
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

// ID is the job's anchor id: "job-" + the slug of its org up to the first
// comma ("Kenya Roads Board, Ministry of…" → "job-kenya-roads-board").
// Skill evidence links point at it; validation keeps it unique.
func (j Job) ID() string {
	org, _, _ := strings.Cut(j.Org, ",")
	return "job-" + Slug(org)
}

// EmploymentTypes are the allowed experience `type` values (as on LinkedIn).
var EmploymentTypes = []string{"Internship", "Volunteer", "Contract", "Freelance", "Full-time", "Part-time"}

// Present is the `end` value of an ongoing role.
const Present = "present"

type Skills struct {
	Eyebrow string       `yaml:"eyebrow"`
	Heading string       `yaml:"heading"`
	Lead    string       `yaml:"lead"`
	Groups  []SkillGroup `yaml:"groups"`
}

type SkillGroup struct {
	Name  string      `yaml:"name"`
	Icon  string      `yaml:"icon"`
	Note  string      `yaml:"note"` // optional mono line under the name
	Items []SkillItem `yaml:"items"`
}

// SkillItem is one skill pill. Evidence (where it was used) is matched on
// Name; Icon is a templates/icons file (a brand logo or a Lucide icon).
type SkillItem struct {
	Name string `yaml:"name"`
	Icon string `yaml:"icon"`
}

// PanelID is the id of the skill's "used in" panel: used-<group>-<skill>.
func (g SkillGroup) PanelID(item SkillItem) string {
	return "used-" + Slug(g.Name) + "-" + Slug(item.Name)
}

type Projects struct {
	Eyebrow string    `yaml:"eyebrow"`
	Heading string    `yaml:"heading"`
	Action  *Action   `yaml:"action"`
	Items   []Project `yaml:"items"`
}

type Project struct {
	Slug        string   `yaml:"slug"`
	Title       string   `yaml:"title"`
	Kind        string   `yaml:"kind"`
	Stack       []string `yaml:"stack"`
	Summary     string   `yaml:"summary"`
	Metric      *Metric  `yaml:"metric"` // optional: "100+" "active users"
	Note        string   `yaml:"note"`   // optional mono line when there's no metric/link
	Links       []Link   `yaml:"links"`
	Screenshot  string   `yaml:"screenshot"`  // optional, under static/; the placeholder shows until it exists
	Placeholder string   `yaml:"placeholder"` // one of Placeholders
}

// Metric is a project's headline number: Value in large Fraunces, then Label.
type Metric struct {
	Value string `yaml:"value"`
	Label string `yaml:"label"`
}

// Placeholders are the abstract phone-screen UIs drawn in CSS until a real
// screenshot exists ("none": no phone at all).
var Placeholders = []string{"journal", "scanner", "none"}

// HasLinks reports whether the project has any links to render.
func (p Project) HasLinks() bool { return len(p.Links) > 0 }

type Education struct {
	Eyebrow string   `yaml:"eyebrow"`
	Heading string   `yaml:"heading"`
	Items   []School `yaml:"items"`
}

// School is one education entry. The first item with Honours renders as the
// degree card; items with Grades render as the quieter school column.
type School struct {
	Degree        string     `yaml:"degree"`
	School        string     `yaml:"school"`
	Date          string     `yaml:"date"`  // "YYYY-MM": a single date (graduation)
	Start         string     `yaml:"start"` // "YYYY": a span of years, with End
	End           string     `yaml:"end"`
	Icon          string     `yaml:"icon"`
	Honours       string     `yaml:"honours"`
	HonoursDetail string     `yaml:"honours_detail"`
	Award         *Award     `yaml:"award"`
	Grades        []Grade    `yaml:"grades"`
	Timeline      *Milestone `yaml:"timeline"` // optional: also a milestone in the career graph
}

type Award struct {
	Title  string `yaml:"title"`
	Detail string `yaml:"detail"`
}

type Grade struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// Milestone is a dated point in the career graph (e.g. graduation).
type Milestone struct {
	Date   string `yaml:"date"` // "YYYY-MM"
	Title  string `yaml:"title"`
	Detail string `yaml:"detail"`
}

// DateText is Date for display: "2025-12" → "Dec 2025".
func (s School) DateText() string {
	t, err := time.Parse("2006-01", s.Date)
	if err != nil {
		return s.Date
	}
	return t.Format("Jan 2006")
}

// Degree is the education item shown as the headline card: the first with
// honours (nil if none).
func (e Education) Degree() *School {
	for i := range e.Items {
		if e.Items[i].Honours != "" {
			return &e.Items[i]
		}
	}
	return nil
}

// Schools are the items shown in the quieter column: those with grades.
func (e Education) Schools() []School {
	var out []School
	for _, it := range e.Items {
		if len(it.Grades) > 0 {
			out = append(out, it)
		}
	}
	return out
}

type Language struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"` // one of LanguageLevels
}

// LanguageLevels is LinkedIn's proficiency scale, lowest first; a level's
// meter shows its position (1–5).
var LanguageLevels = []string{"Elementary", "Limited working", "Professional working", "Full professional", "Native"}

// Meter is the level's position on LanguageLevels (1–5), 0 if unknown.
func (l Language) Meter() int {
	for i, v := range LanguageLevels {
		if v == l.Level {
			return i + 1
		}
	}
	return 0
}

// MeterLabel is the meter's accessible name: "Level 3 of 5".
func (l Language) MeterLabel() string {
	return fmt.Sprintf("Level %d of %d", l.Meter(), len(LanguageLevels))
}

// MeterBars is one bool per bar of the meter: true = lit.
func (l Language) MeterBars() []bool {
	bars := make([]bool, len(LanguageLevels))
	for i := range l.Meter() {
		bars[i] = true
	}
	return bars
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
	PDF              string    `yaml:"pdf"` // visually hidden suffix of links to a PDF: "(PDF)"
	CTAProjects      string    `yaml:"cta_projects"`
	CTAResume        string    `yaml:"cta_resume"`
	Current          string    `yaml:"current"`
	LanguagesHeading string    `yaml:"languages_heading"`
	SocialLabel      string    `yaml:"social_label"`
	Career           CareerUI  `yaml:"career"`
	About            AboutUI   `yaml:"about"`
	Skills           SkillsUI  `yaml:"skills"`
	Projects         ProjectUI `yaml:"projects"`
	Contact          ContactUI `yaml:"contact"`
	Footer           FooterUI  `yaml:"footer"`
	NotFound         NotFound  `yaml:"not_found"`
}

// AboutUI: the derived fact's label.
type AboutUI struct {
	Currently string `yaml:"currently"` // label of the derived "Currently" fact
}

// SkillsUI words the skill-evidence panels: "{skill}, {used in|used at|shown at}".
type SkillsUI struct {
	UsedIn  string `yaml:"used_in"`  // all evidence is projects
	UsedAt  string `yaml:"used_at"`  // all evidence is paid work
	ShownAt string `yaml:"shown_at"` // mixed, or includes volunteer roles
	Project string `yaml:"project"`  // kind label of project evidence
	Place   string `yaml:"place"`    // a skill button's hidden count: "used in 1 place"
	Places  string `yaml:"places"`   // "used in 2 places"
}

// ProjectUI labels the project cards.
type ProjectUI struct {
	StackLabel    string `yaml:"stack_label"`    // accessible name of the tag list ("Built with")
	ScreenshotAlt string `yaml:"screenshot_alt"` // alt prefix: "Screenshot of" + title
	Journal       struct {
		Title string `yaml:"title"`
	} `yaml:"journal"` // placeholder phone UI copy (decorative)
	Scanner struct {
		Title  string `yaml:"title"`
		Status string `yaml:"status"`
	} `yaml:"scanner"`
}

// ContactUI: the copy-email button and the local-time line.
type ContactUI struct {
	Copy           string `yaml:"copy"`            // button label
	Copied         string `yaml:"copied"`          // after a successful copy
	Fallback       string `yaml:"fallback"`        // clipboard blocked: text selected instead
	CopiedStatus   string `yaml:"copied_status"`   // announced after a copy
	FallbackStatus string `yaml:"fallback_status"` // announced when the text was selected
	TimeBefore     string `yaml:"time_before"`     // "It's" {time}
	TimeAfter      string `yaml:"time_after"`      // {time} "here" ({zone})
}

// FooterUI is the footer's copy.
type FooterUI struct {
	Built string `yaml:"built"`  // "Built with Golang"
	ToTop string `yaml:"to_top"` // "Back to top"
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
	// Announced through the page's live region when the legend filter changes.
	ShowingAll        string `yaml:"showing_all"`        // "Showing all roles"
	Showing           string `yaml:"showing"`            // "Showing {kind} roles" ({kind}: the type, lower-cased)
	ShowingMilestones string `yaml:"showing_milestones"` // "Showing milestones"
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
	// Announced through the page's live region when the theme changes.
	// {theme} in AnnounceSystem is the resolved theme's label, lower-cased.
	AnnounceDark   string `yaml:"announce_dark"`   // "Dark theme"
	AnnounceLight  string `yaml:"announce_light"`  // "Light theme"
	AnnounceSystem string `yaml:"announce_system"` // "System theme, currently {theme}"
}

// NotFound is the 404 page's copy: the aside title, the body and the two buttons.
type NotFound struct {
	Title      string `yaml:"title"`
	Body       string `yaml:"body"`
	Home       string `yaml:"home"`       // primary button, to /
	Experience string `yaml:"experience"` // ghost button, to /#experience
}

// Slug lower-cases s and joins its runs of letters and digits with "-":
// "Git & GitHub" → "git-github", "IEEE Tech4Good" → "ieee-tech4good".
func Slug(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if dash && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			dash = false
		} else {
			dash = true
		}
	}
	return b.String()
}
