// Package seo generates what search engines and link previews read: the
// Person JSON-LD, robots.txt, sitemap.xml and the web app manifest. It is
// pure (no file access); internal/build writes the files and internal/render
// embeds the JSON-LD.
package seo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"slices"
	"strings"
	"time"

	"github.com/kmuchina/portfolio/internal/site"
)

// Thing is a schema.org node with a type and a name.
type Thing struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

// PostalAddress is schema.org's PostalAddress (city and country only).
type PostalAddress struct {
	Type     string `json:"@type"`
	Locality string `json:"addressLocality"`
	Country  string `json:"addressCountry"`
}

// Person is the page's one JSON-LD node. There is deliberately no
// telephone field: no phone number is published anywhere.
type Person struct {
	Context    string        `json:"@context"`
	Type       string        `json:"@type"`
	Name       string        `json:"name"`
	GivenName  string        `json:"givenName"`
	Additional string        `json:"additionalName,omitempty"`
	FamilyName string        `json:"familyName"`
	JobTitle   string        `json:"jobTitle"`
	URL        string        `json:"url"`
	Image      string        `json:"image"`
	Email      string        `json:"email"`
	Address    PostalAddress `json:"address"`
	AlumniOf   *Thing        `json:"alumniOf,omitempty"`
	WorksFor   *Thing        `json:"worksFor,omitempty"`
	KnowsAbout []string      `json:"knowsAbout,omitempty"`
	SameAs     []string      `json:"sameAs,omitempty"`
}

// HomeURL is site.url with exactly one trailing slash: "https://kmuchina.com/".
func HomeURL(siteURL string) string { return strings.TrimRight(siteURL, "/") + "/" }

// Abs turns a root-relative path ("/img/a.jpg") into an absolute URL.
func Abs(siteURL, p string) string { return strings.TrimRight(siteURL, "/") + "/" + strings.TrimLeft(p, "/") }

// NewPerson builds the Person node from the content. image is the absolute
// URL of the profile photo. worksFor is the current role's organisation (the
// experience item whose end is "present"), alumniOf the degree's school,
// knowsAbout the skills of the groups named in seo.knows_about, and sameAs
// every contact link.
func NewPerson(s *site.Site, image string) Person {
	p := Person{
		Context:    "https://schema.org",
		Type:       "Person",
		Name:       s.Person.FullName(),
		GivenName:  s.Person.FirstName,
		Additional: s.Person.MiddleName,
		FamilyName: s.Person.LastName,
		JobTitle:   s.Person.Role,
		URL:        HomeURL(s.Meta.URL),
		Image:      image,
		Email:      s.Contact.MailtoHref(),
		Address:    PostalAddress{Type: "PostalAddress", Locality: s.SEO.Address.Locality, Country: s.SEO.Address.Country},
	}
	if d := s.Education.Degree(); d != nil {
		p.AlumniOf = &Thing{Type: "CollegeOrUniversity", Name: d.School}
	}
	for _, j := range s.Experience.Items {
		if j.End == site.Present {
			org, _, _ := strings.Cut(j.Org, ",")
			p.WorksFor = &Thing{Type: "Organization", Name: strings.TrimSpace(org)}
			break
		}
	}
	for _, g := range s.Skills.Groups {
		if slices.Contains(s.SEO.KnowsAbout, g.Name) {
			for _, it := range g.Items {
				p.KnowsAbout = append(p.KnowsAbout, it.Name)
			}
		}
	}
	for _, l := range s.Contact.Links {
		p.SameAs = append(p.SameAs, l.URL)
	}
	return p
}

// JSONLD marshals v for a <script type="application/ld+json">.
// encoding/json escapes <, > and & as <…, so the payload can never
// close the script element.
func JSONLD(v any) ([]byte, error) { return json.Marshal(v) }

// Robots is robots.txt: everything is allowed except the dev styleguide,
// and the sitemap is announced.
func Robots(siteURL string) []byte {
	return []byte("User-agent: *\nAllow: /\nDisallow: /styleguide.html\n\nSitemap: " + Abs(siteURL, "sitemap.xml") + "\n")
}

// Sitemap lists only the home page, with lastmod = the build date.
func Sitemap(siteURL string, now time.Time) ([]byte, error) {
	type url struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod"`
	}
	type urlset struct {
		XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
		URLs    []url    `xml:"url"`
	}
	body, err := xml.MarshalIndent(urlset{URLs: []url{{Loc: HomeURL(siteURL), LastMod: now.Format("2006-01-02")}}}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(append([]byte(xml.Header), body...), '\n'), nil
}

// Icon is one web manifest icon.
type Icon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
	Type  string `json:"type"`
}

// Manifest is a minimal site.webmanifest: no service worker, display
// "browser" (it's a website, not an installable app).
type Manifest struct {
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	StartURL        string `json:"start_url"`
	Display         string `json:"display"`
	ThemeColor      string `json:"theme_color"`
	BackgroundColor string `json:"background_color"`
	Icons           []Icon `json:"icons"`
}

// JSON is the manifest file's contents.
func (m Manifest) JSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	err := enc.Encode(m)
	return buf.Bytes(), err
}
