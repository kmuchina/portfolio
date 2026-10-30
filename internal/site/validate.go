package site

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// retiredEmail must never appear anywhere in the content.
const retiredEmail = "kmuchina22"

// SectionIDs are the anchor ids of the page sections, in page order.
var SectionIDs = []string{"top", "about", "experience", "skills", "projects", "education", "contact"}

var (
	emailRe  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	slugRe   = regexp.MustCompile(`^[a-z0-9-]+$`)
	monthRe  = regexp.MustCompile(`^[0-9]{4}-(0[1-9]|1[0-2])$`)
	internRe = regexp.MustCompile(`(?i)\bintern(s|ship)?\b`) // "Intern", "internship"; not "International"
)

// Problems is every validation failure, reported together.
type Problems []string

func (p Problems) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "content has %d problem(s):", len(p))
	for i, msg := range p {
		fmt.Fprintf(&b, "\n  %d. %s", i+1, msg)
	}
	return b.String()
}

// Validate checks s against the files under staticDir and the icon SVGs in
// iconDir. It collects ALL problems and returns them as one Problems error.
// Warnings (e.g. a missing project screenshot) never fail the build.
func Validate(s *Site, staticDir, iconDir string) (warnings []string, err error) {
	var probs Problems
	add := func(format string, args ...any) { probs = append(probs, fmt.Sprintf(format, args...)) }

	required := []struct{ field, value string }{
		{"site.url", s.Meta.URL},
		{"site.title", s.Meta.Title},
		{"site.description", s.Meta.Description},
		{"site.lang", s.Meta.Lang},
		{"person.first_name", s.Person.FirstName},
		{"person.middle_name", s.Person.MiddleName},
		{"person.last_name", s.Person.LastName},
		{"person.role", s.Person.Role},
		{"person.summary", s.Person.Summary},
		{"person.photo_alt", s.Person.PhotoAlt},
		{"contact.email", s.Contact.Email},
		{"contact.heading", s.Contact.Heading},
		{"about.heading", s.About.Heading},
		{"experience.heading", s.Experience.Heading},
		{"skills.heading", s.Skills.Heading},
		{"projects.heading", s.Projects.Heading},
		{"education.heading", s.Education.Heading},
		{"ui.languages_heading", s.UI.LanguagesHeading},
		{"ui.nav_label", s.UI.NavLabel},
		{"ui.home_label", s.UI.HomeLabel},
		{"ui.menu.label", s.UI.Menu.Label},
		{"ui.menu.open", s.UI.Menu.Open},
		{"ui.menu.close", s.UI.Menu.Close},
		{"ui.theme.label", s.UI.Theme.Label},
		{"ui.theme.menu_label", s.UI.Theme.MenuLabel},
		{"ui.career.highlight", s.UI.Career.Highlight},
		{"ui.career.legend_label", s.UI.Career.LegendLabel},
		{"ui.career.all", s.UI.Career.All},
		{"ui.career.milestone", s.UI.Career.Milestone},
		{"ui.career.present", s.UI.Career.Present},
		{"ui.career.skills_label", s.UI.Career.SkillsLabel},
	}
	for _, r := range required {
		if strings.TrimSpace(r.value) == "" {
			add("%s is required", r.field)
		}
	}

	if s.Contact.Email != "" && !emailRe.MatchString(s.Contact.Email) {
		add("contact.email %q is not a valid email address", s.Contact.Email)
	}
	if s.Meta.ShowPhone && strings.TrimSpace(s.Contact.Phone) == "" {
		add("site.show_phone is true but contact.phone is empty")
	}

	// The retired address must not appear in any loaded field.
	if dump, err := yaml.Marshal(s); err == nil && strings.Contains(string(dump), retiredEmail) {
		add("content contains the retired address %s@…; use muchina537@gmail.com", retiredEmail)
	}

	checkURL := func(field, raw string) {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			add("%s %q must be an absolute https:// URL", field, raw)
		}
	}
	if s.Meta.URL != "" {
		checkURL("site.url", s.Meta.URL)
	}

	checkFile := func(field, rel string) {
		if rel == "" {
			add("%s is required", field)
			return
		}
		if _, err := os.Stat(filepath.Join(staticDir, filepath.FromSlash(rel))); err != nil {
			add("%s %q not found under %s/", field, rel, staticDir)
		}
	}
	checkFile("person.photo", s.Person.Photo)
	// The résumé is optional: empty hides the hero download button.
	if s.Person.ResumePDF != "" {
		checkFile("person.resume_pdf", s.Person.ResumePDF)
	}

	checkAction := func(field string, a *Action) {
		if a == nil {
			return
		}
		if strings.TrimSpace(a.Label) == "" {
			add("%s.label is required", field)
		}
		switch {
		case a.URL == "":
			add("%s.url is required", field)
		case a.IsExternal():
			checkURL(field+".url", a.URL)
		default:
			checkFile(field+".url", a.AssetPath())
		}
	}
	checkAction("experience.action", s.Experience.Action)
	checkMonth := func(field, v string) {
		if !monthRe.MatchString(v) {
			add("%s %q must be YYYY-MM", field, v)
		}
	}
	if len(s.Experience.Items) == 0 {
		add("experience.items needs at least one entry")
	}
	for i, j := range s.Experience.Items {
		f := fmt.Sprintf("experience.items[%d] (%s)", i, j.Title)
		for _, r := range []struct{ field, value string }{
			{"title", j.Title}, {"org", j.Org}, {"type", j.Type}, {"icon", j.Icon}, {"location", j.Location},
		} {
			if strings.TrimSpace(r.value) == "" {
				add("%s.%s is required", f, r.field)
			}
		}
		if internRe.MatchString(j.Title) {
			add("%s.title must not contain \"intern\": the employment type says that (type: Internship)", f)
		}
		if j.Type != "" && !slices.Contains(EmploymentTypes, j.Type) {
			add("%s.type %q must be one of %s", f, j.Type, strings.Join(EmploymentTypes, " | "))
		}
		checkMonth(f+".start", j.Start)
		if j.End != Present {
			checkMonth(f+".end", j.End)
			if monthRe.MatchString(j.Start) && monthRe.MatchString(j.End) && j.End < j.Start {
				add("%s.end %s is before start %s", f, j.End, j.Start)
			}
		}
		if len(j.Bullets) == 0 {
			add("%s.bullets needs at least one bullet", f)
		}
	}
	for i, e := range s.Education.Items {
		if m := e.Timeline; m != nil {
			f := fmt.Sprintf("education.items[%d].timeline", i)
			checkMonth(f+".date", m.Date)
			if strings.TrimSpace(m.Title) == "" {
				add("%s.title is required", f)
			}
		}
	}
	checkAction("projects.action", s.Projects.Action)

	checkIcon := func(field, name string) {
		if name == "" {
			return
		}
		if _, err := os.Stat(filepath.Join(iconDir, name+".svg")); err != nil {
			add("%s: icon %q has no %s/%s.svg", field, name, iconDir, name)
		}
	}

	for i, l := range s.Contact.Links {
		f := fmt.Sprintf("contact.links[%d] (%s)", i, l.Label)
		checkURL(f+".url", l.URL)
		if l.Icon == "" {
			add("%s.icon is required", f)
		}
		checkIcon(f, l.Icon)
	}
	for i, it := range s.About.Focus {
		checkIcon(fmt.Sprintf("about.focus[%d]", i), it.Icon)
	}
	for i, it := range s.About.Facts {
		checkIcon(fmt.Sprintf("about.facts[%d]", i), it.Icon)
	}
	for i, j := range s.Experience.Items {
		checkIcon(fmt.Sprintf("experience.items[%d]", i), j.Icon)
	}
	for i, g := range s.Skills.Groups {
		checkIcon(fmt.Sprintf("skills.groups[%d]", i), g.Icon)
	}

	seen := map[string]bool{}
	for i, p := range s.Projects.Items {
		f := fmt.Sprintf("projects.items[%d]", i)
		switch {
		case !slugRe.MatchString(p.Slug):
			add("%s.slug %q must match %s", f, p.Slug, slugRe)
		case seen[p.Slug]:
			add("%s.slug %q is not unique", f, p.Slug)
		}
		seen[p.Slug] = true
		for j, l := range p.Links {
			lf := fmt.Sprintf("%s.links[%d]", f, j)
			checkURL(lf+".url", l.URL)
			checkIcon(lf, l.Icon)
		}
		if p.Image != "" {
			if _, err := os.Stat(filepath.Join(staticDir, filepath.FromSlash(p.Image))); err != nil {
				warnings = append(warnings, fmt.Sprintf("%s (%s): image %q not found; it will be omitted", f, p.Slug, p.Image))
			} else if strings.TrimSpace(p.ImageAlt) == "" {
				add("%s.image_alt is required when the image exists", f)
			}
		}
	}

	ids := map[string]bool{}
	for _, id := range SectionIDs {
		ids["#"+id] = true
	}
	if len(s.UI.Nav) == 0 {
		add("ui.nav needs at least one link")
	}
	if s.UI.Theme.Label != "" && s.UI.Theme.Label == s.UI.Theme.MenuLabel {
		add("ui.theme.label and ui.theme.menu_label must differ (two toggles share a page)")
	}
	for i, n := range s.UI.Nav {
		if !ids[n.Href] {
			add("ui.nav[%d] (%s) href %q is not a section anchor (%s)", i, n.Label, n.Href, strings.Join(SectionIDs, ", "))
		}
	}

	if len(probs) > 0 {
		return warnings, probs
	}
	return warnings, nil
}
