package site

import (
	"fmt"
	"image"
	_ "image/png" // DecodeConfig of the share image
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

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
	yearRe   = regexp.MustCompile(`^[0-9]{4}$`)
	internRe = regexp.MustCompile(`(?i)\bintern(s|ship)?\b`) // "Intern", "internship"; not "International"
	localeRe  = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)
	countryRe = regexp.MustCompile(`^[A-Z]{2}$`)
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
		{"ui.career.showing_all", s.UI.Career.ShowingAll},
		{"ui.career.showing", s.UI.Career.Showing},
		{"ui.career.showing_milestones", s.UI.Career.ShowingMilestones},
		{"ui.theme.announce_dark", s.UI.Theme.AnnounceDark},
		{"ui.theme.announce_light", s.UI.Theme.AnnounceLight},
		{"ui.theme.announce_system", s.UI.Theme.AnnounceSystem},
		{"ui.pdf", s.UI.PDF},
		{"contact.timezone.iana", s.Contact.Timezone.IANA},
		{"contact.timezone.label", s.Contact.Timezone.Label},
		{"ui.about.currently", s.UI.About.Currently},
		{"ui.skills.used_in", s.UI.Skills.UsedIn},
		{"ui.skills.used_at", s.UI.Skills.UsedAt},
		{"ui.skills.shown_at", s.UI.Skills.ShownAt},
		{"ui.skills.project", s.UI.Skills.Project},
		{"ui.skills.place", s.UI.Skills.Place},
		{"ui.skills.places", s.UI.Skills.Places},
		{"ui.projects.stack_label", s.UI.Projects.StackLabel},
		{"ui.projects.screenshot_alt", s.UI.Projects.ScreenshotAlt},
		{"ui.contact.copy", s.UI.Contact.Copy},
		{"ui.contact.copied", s.UI.Contact.Copied},
		{"ui.contact.fallback", s.UI.Contact.Fallback},
		{"ui.contact.copied_status", s.UI.Contact.CopiedStatus},
		{"ui.contact.fallback_status", s.UI.Contact.FallbackStatus},
		{"ui.footer.built", s.UI.Footer.Built},
		{"ui.footer.to_top", s.UI.Footer.ToTop},
		{"ui.not_found.title", s.UI.NotFound.Title},
		{"ui.not_found.body", s.UI.NotFound.Body},
		{"ui.not_found.home", s.UI.NotFound.Home},
		{"ui.not_found.experience", s.UI.NotFound.Experience},
		{"seo.locale", s.SEO.Locale},
		{"seo.short_name", s.SEO.ShortName},
		{"seo.image_alt", s.SEO.ImageAlt},
		{"seo.address.locality", s.SEO.Address.Locality},
		{"seo.address.country", s.SEO.Address.Country},
	}
	for _, r := range required {
		if strings.TrimSpace(r.value) == "" {
			add("%s is required", r.field)
		}
	}

	if s.Contact.Email != "" && !emailRe.MatchString(s.Contact.Email) {
		add("contact.email %q is not a valid email address", s.Contact.Email)
	}
	if tz := s.Contact.Timezone.IANA; tz != "" {
		if _, err := time.LoadLocation(tz); err != nil {
			add("contact.timezone.iana %q is not a time zone: %v", tz, err)
		}
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
	jobIDs := map[string]bool{}
	for i, j := range s.Experience.Items {
		if id := j.ID(); jobIDs[id] {
			add("experience.items[%d] (%s): anchor id %q is not unique (it comes from org)", i, j.Title, id)
		} else {
			jobIDs[id] = true
		}
	}
	degrees := 0
	for i, e := range s.Education.Items {
		f := fmt.Sprintf("education.items[%d] (%s)", i, e.Degree)
		if strings.TrimSpace(e.Degree) == "" || strings.TrimSpace(e.School) == "" {
			add("%s: degree and school are required", f)
		}
		if e.Date != "" {
			checkMonth(f+".date", e.Date)
		}
		if e.Start != "" || e.End != "" {
			if !yearRe.MatchString(e.Start) || !yearRe.MatchString(e.End) {
				add("%s: start %q and end %q must both be YYYY", f, e.Start, e.End)
			} else if e.End < e.Start {
				add("%s.end %s is before start %s", f, e.End, e.Start)
			}
		}
		switch {
		case e.Honours != "":
			degrees++
		case len(e.Grades) == 0:
			add("%s needs honours (the degree card) or grades (the school column); otherwise it isn't shown", f)
		}
		for j, g := range e.Grades {
			if strings.TrimSpace(g.Label) == "" || strings.TrimSpace(g.Value) == "" {
				add("%s.grades[%d]: label and value are required", f, j)
			}
		}
		if a := e.Award; a != nil && strings.TrimSpace(a.Title) == "" {
			add("%s.award.title is required", f)
		}
		if m := e.Timeline; m != nil {
			checkMonth(f+".timeline.date", m.Date)
			if strings.TrimSpace(m.Title) == "" {
				add("%s.timeline.title is required", f)
			}
		}
	}
	if degrees > 1 {
		add("education: only the first item with honours is shown as the degree card (%d have honours)", degrees)
	}
	for i, l := range s.Languages {
		if !slices.Contains(LanguageLevels, l.Level) {
			add("languages[%d] (%s).level %q must be one of %s", i, l.Name, l.Level, strings.Join(LanguageLevels, " | "))
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
		f := fmt.Sprintf("about.focus[%d]", i)
		if strings.TrimSpace(it.Label) == "" || it.Icon == "" {
			add("%s: label and icon are required", f)
		}
		checkIcon(f, it.Icon)
	}
	for i, it := range s.About.Facts {
		if strings.TrimSpace(it.Label) == "" || strings.TrimSpace(it.Value) == "" {
			add("about.facts[%d]: label and value are required", i)
		}
	}
	for i, e := range s.Education.Items {
		checkIcon(fmt.Sprintf("education.items[%d]", i), e.Icon)
	}
	for i, j := range s.Experience.Items {
		checkIcon(fmt.Sprintf("experience.items[%d]", i), j.Icon)
	}
	panelIDs := map[string]bool{}
	for i, g := range s.Skills.Groups {
		gf := fmt.Sprintf("skills.groups[%d] (%s)", i, g.Name)
		if strings.TrimSpace(g.Name) == "" || g.Icon == "" {
			add("%s: name and icon are required", gf)
		}
		checkIcon(gf, g.Icon)
		for j, it := range g.Items {
			f := fmt.Sprintf("%s.items[%d] (%s)", gf, j, it.Name)
			if strings.TrimSpace(it.Name) == "" {
				add("%s.name is required", f)
			}
			if it.Icon == "" {
				add("%s.icon is required (every skill pill has an icon)", f)
			}
			checkIcon(f, it.Icon)
			if id := g.PanelID(it); panelIDs[id] {
				add("%s: panel id %q is not unique", f, id)
			} else {
				panelIDs[id] = true
			}
		}
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
			if strings.TrimSpace(l.Label) == "" {
				add("%s.label is required", lf)
			}
			checkURL(lf+".url", l.URL)
			checkIcon(lf, l.Icon)
		}
		if m := p.Metric; m != nil && (strings.TrimSpace(m.Value) == "" || strings.TrimSpace(m.Label) == "") {
			add("%s.metric: value and label are required", f)
		}
		if !slices.Contains(Placeholders, p.Placeholder) {
			add("%s.placeholder %q must be one of %s", f, p.Placeholder, strings.Join(Placeholders, " | "))
		}
		if p.Screenshot != "" {
			if _, err := os.Stat(filepath.Join(staticDir, filepath.FromSlash(p.Screenshot))); err != nil {
				what := "the " + p.Placeholder + " placeholder shows instead"
				if p.Placeholder == "none" {
					what = "no phone is shown"
				}
				warnings = append(warnings, fmt.Sprintf("%s (%s): screenshot %q not found; %s", f, p.Slug, p.Screenshot, what))
			}
		}
	}

	// SEO and sharing.
	if n := utf8.RuneCountInString(s.Meta.Description); s.Meta.Description != "" && (n < DescriptionMin || n > DescriptionMax) {
		warnings = append(warnings, fmt.Sprintf("site.description is %d characters; search results show %d–%d in full", n, DescriptionMin, DescriptionMax))
	}
	if s.SEO.Locale != "" && !localeRe.MatchString(s.SEO.Locale) {
		add("seo.locale %q must look like en_KE (language_COUNTRY)", s.SEO.Locale)
	}
	if c := s.SEO.Address.Country; c != "" && !countryRe.MatchString(c) {
		add("seo.address.country %q must be a two-letter country code (KE)", c)
	}
	checkFile("seo.image", s.SEO.Image)
	if s.SEO.Image != "" {
		if f, err := os.Open(filepath.Join(staticDir, filepath.FromSlash(s.SEO.Image))); err == nil {
			cfg, format, err := image.DecodeConfig(f)
			f.Close()
			if err != nil || format != "png" || cfg.Width != ShareImageW || cfg.Height != ShareImageH {
				add("seo.image %q must be a %d × %d PNG", s.SEO.Image, ShareImageW, ShareImageH)
			}
		}
	}
	if len(s.SEO.KnowsAbout) == 0 {
		add("seo.knows_about needs at least one skill group")
	}
	for i, name := range s.SEO.KnowsAbout {
		if !slices.ContainsFunc(s.Skills.Groups, func(g SkillGroup) bool { return g.Name == name }) {
			add("seo.knows_about[%d] %q is not a skills.groups name", i, name)
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
