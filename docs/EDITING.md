# Editing the site

A plain guide to changing kmuchina.com yourself. Almost everything lives in
one file: **`content/site.yaml`**. The templates, CSS and Go code read from
it, so you rarely need to touch anything else.

## The routine for every change

1. Start the preview: `make serve`, then open <http://localhost:8080>.
   Every refresh rebuilds the site, so save the YAML and reload.
2. If the build has a problem, the browser shows a numbered list of
   everything that's wrong (see [What the build errors mean](#what-the-build-errors-mean)).
3. When it looks right, run **`make ci`**. It runs exactly what GitHub runs
   before deploying.
4. Commit and push to `main`. The site deploys itself in about a minute.
   **Pushing to `main` publishes.**

YAML tips: indentation is two spaces (never tabs); put text in double quotes;
inside double quotes write `\"` for a quote character.

---

## Adding a new job

Jobs live under `experience: items:`. The order in the file doesn't matter:
the career graph sorts by date, newest first.

```yaml
    - title: "Software Engineer"            # never include "intern": the type says that
      org: "Acme Ltd, Nairobi"              # text after the first comma is not used in the job's link id
      type: "Full-time"
      icon: "briefcase"                     # any icon file in templates/icons/ (no company logos)
      location: "Nairobi, Kenya"
      start: "2027-01"                      # YYYY-MM
      end: "present"                        # YYYY-MM, or "present" for the job you're in now
      bullets:
        - "Built the payments API, cutting checkout time by **40%**."   # **bold** marks a number
        - "Mentored **3** junior developers."
      skills: ["Kotlin", "REST APIs"]       # words that appear in the bullets; nothing invented
```

**`type` picks the lane and the colour.** It must be one of these, exactly
as written:

| type | Lane in the graph | Colour |
|---|---|---|
| `Internship` | work (left) | blue |
| `Full-time`, `Part-time` | work (left) | ink (white in dark, navy in light) |
| `Contract` | work (left) | amber |
| `Freelance` | work (left) | violet |
| `Volunteer` | volunteer (right) | green, diamond node |

Jobs in the same lane that overlap in time share the lane. The legend under
"Where I've worked" adds a button for each type you use, by itself.

**`end: "present"` makes it the current role:** a larger node with a
gently breathing halo, the line fading upward above it, a live duration
("3 mos"), and it becomes the **"Currently"** fact in About and the
`worksFor` organisation that Google sees. Only one job should be `present`:
when you start a new job, give the old one its real `end` month.

**Skills on a job become evidence.** If a `skills:` entry has the same name
as a skill pill (see below), that pill gets a number and a "used at" link
to this job.

## Adding a project and its screenshot

Projects live under `projects: items:`.

```yaml
    - slug: "budget-buddy"                  # lowercase letters, digits and dashes; unique
      title: "Budget Buddy"
      kind: "Android app"
      stack: ["Kotlin", "Jetpack Compose", "Room"]
      summary: "One or two sentences on what it does and what was hard."
      metric: { value: "500+", label: "downloads" }       # optional
      links:                                              # optional; the first link makes the whole card clickable
        - { label: "See it on Google Play", url: "https://play.google.com/store/apps/details?id=…" }
      screenshot: "img/projects/budget-buddy.png"         # optional
      placeholder: "journal"                              # journal | scanner | none: shown until the screenshot exists
```

Use `note: "Runs fully on-device"` instead of `metric` and `links` when
there's nothing to link to.

**The screenshot:**
- Save it as **`static/img/projects/<slug>.png`** (or `.jpg`), the same
  path as `screenshot:`.
- **Portrait**, ideally **1080 × 2220** (a phone screen). It's shown inside
  the phone frame, about 184px wide, so anything at least 720px wide is plenty.
- You don't make other sizes yourself: the build writes 360, 540 and 720px
  versions (JPEG, plus WebP on GitHub) and never ships your original file.
- Until the file exists, the build prints a warning and shows the drawn
  placeholder. That's fine.

**Stack words become evidence**: a stack entry with the same name as a
skill pill links that pill to this project ("used in").

## Adding a skill and its icon

Skills live under `skills: groups:`. Every pill needs an icon.

```yaml
    - name: "Programming"
      icon: "code"
      items:
        - { name: "TypeScript", icon: "typescript" }
```

**The icon** must be a file in `templates/icons/` (the name without `.svg`).
Look there first: about 80 are already in. To add one:

- **A technology logo:** from Simple Icons (CC0). In a scratch folder run
  `npm pack simple-icons`, unpack it, and copy `icons/<name>.svg`. Never
  draw a logo yourself or take an old version of a removed logo.
- **Anything else:** from Lucide (ISC): `npm pack lucide-static`, then
  `icons/<name>.svg`.
- Put it in `templates/icons/<name>.svg` on **one line**, with its `viewBox`
  and without `width`, `height`, `class` or `stroke-width`. Lucide files
  keep `fill="none" stroke="currentColor"` and round caps and joins; Simple
  Icons files use `fill="currentColor"`. Copy the shape of an existing file.

**Evidence appears by itself.** The build matches skill names against
every project's `stack` and every job's `skills`. Case and punctuation are
ignored ("Node.js" = "nodejs"), but the words must match: "C++" is not "C".
A matching skill becomes a button with a count, and its panel links to the
projects and jobs.

## Updating the résumé PDF

1. Before replacing it, check the new PDF has no phone number and no
   referees' details:
   ```sh
   pdftotext new-resume.pdf - | grep -i -e 703 -e +254 -e referee -e kmuchina22
   ```
   No output means it's clean.
2. Save it as **`static/Kennedy_Muchina_Resume.pdf`**, replacing the old one.
   With the same name, nothing else changes.
3. If you change the file name, change it in two places in `site.yaml`:
   `person.resume_pdf` and `experience.action.url`.

## Changing the share image

The picture LinkedIn, WhatsApp and X show when someone shares the link is
`static/og.png` (1200 × 630). It's drawn from `design/og.html`.

1. Edit `design/og.html`. Open it in a browser to preview (it reads the
   real fonts and colours).
2. Run **`make og`**. It needs Playwright. If it says it can't find it,
   either install it (`npm install -g playwright && npx playwright install chromium`)
   or point at an existing copy:
   `PLAYWRIGHT_MODULE=/path/to/node_modules/playwright-core make og`.
   It also re-renders `favicon.ico` and `apple-touch-icon.png` from
   `design/icons.html`, and prints the file sizes.
3. Keep `og.png` at **200 KB or less**. Commit the PNGs.
4. Social sites cache previews. After deploying, paste the URL into
   [LinkedIn Post Inspector](https://www.linkedin.com/post-inspector/) to
   refresh LinkedIn's copy.

The alt text and the other share details live under `seo:` in `site.yaml`.

---

## What the build errors mean

The build lists **every** problem at once, numbered (`content has 3
problem(s): 1. … 2. …`), with the item's name in brackets so you can find
it. Fix them from the top.

**Unknown field**
```
build failed: content/site.yaml: yaml: unmarshal errors:
  line 84: field titel not found in type site.Job
```
A key is misspelt (`titel`), or it isn't a field the site knows. Check the
spelling against an existing entry. A brand-new kind of field needs a code
change first.

**Bad URL**
```
contact.links[1] (GitHub).url "github.com/kmuchina" must be an absolute https:// URL
```
Links must start with `https://`.

**Missing icon**
```
skills.groups[0] (Programming).items[8] (TypeScript): icon "typescript" has no templates/icons/typescript.svg
```
The icon file doesn't exist. Add it (see above) or use one that does.

**"intern" in a title**
```
experience.items[0] (Software Intern).title must not contain "intern": the employment type says that (type: Internship)
```
Write the role ("Software Engineer") and put `type: "Internship"`.

**Wrong type or date**
```
experience.items[0] (…).type "Fulltime" must be one of Internship | Volunteer | Contract | Freelance | Full-time | Part-time
experience.items[0] (…).start "Jan 2027" must be YYYY-MM
```

**Phone number detected** (from `make ci` / `make test`)
```
dist/index.html contains "+254": no phone number on the site
```
A phone number has crept in somewhere: the YAML, the résumé PDF or a
template. The site never shows one; remove it.

**Retired email**
```
content contains the retired address kmuchina22@…; use muchina537@gmail.com
```

**Share image**
```
seo.image "og.png" must be a 1200 × 630 PNG
```
Re-render it with `make og`.

### Warnings (the build still succeeds)

- `screenshot "img/projects/x.png" not found; the journal placeholder shows instead`:
  expected until you add the screenshot.
- `cwebp not found on PATH: WebP images skipped`: only on your machine.
  GitHub installs `cwebp`, so the live site has WebP. Locally,
  `sudo apt install webp` silences it.
- `site.description is 132 characters; search results show 150–160 in full`:
  lengthen or shorten `site.description`.
