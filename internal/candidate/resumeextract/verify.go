package resumeextract

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// This file is the gate a parse made OUTSIDE the server passes before it is stored: a
// candidate's own agent (Claude Code, Codex) reads their CV and submits the structure the
// in-process Extractor would otherwise have produced. The server cannot see that model's
// reasoning, so it checks the result against the one thing it does hold — the text of the
// uploaded file — and keeps only what that text actually says.
//
// The rule is transcription, the same rule the Extractor's own prompt states: a parse
// restates the CV, it does not write it. A value the text does not contain is not a fact
// about the candidate, however plausible, and stored here it would be seeded into every CV
// they render and banked as cv_import evidence the CV gate trusts. So every textual value is
// checked by containment, and what fails is dropped and reported rather than kept or
// rejected wholesale — one paraphrased bullet must not cost the candidate their education.

// sourceNoiseRE folds everything that is not a letter or a digit into a separator: the same
// normalisation experience.ClaimKey applies to a banked claim (case, punctuation and spacing
// do not change what a bullet says; numbers do, and are kept).
var sourceNoiseRE = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// sourceKey normalises a value for the containment check. NFKC first, because PDF text
// layers carry presentation forms — the "ﬁ" ligature, full-width digits — that print as the
// letters a model types and compare as different runes.
func sourceKey(s string) string {
	folded := strings.ToLower(norm.NFKC.String(s))
	return strings.TrimSpace(sourceNoiseRE.ReplaceAllString(folded, " "))
}

// sourceText is a CV's text, normalised once, answering whether a value appears in it as a
// run of whole words. Padding with a separator on both sides is what makes "java" not match
// inside "javascript", and "5" not match inside "2015".
type sourceText struct{ padded string }

func newSourceText(text string) sourceText {
	return sourceText{padded: " " + sourceKey(text) + " "}
}

func (t sourceText) has(value string) bool {
	k := sourceKey(value)
	return k != "" && strings.Contains(t.padded, " "+k+" ")
}

// Unsupported names one value KeepSupported removed, for the caller to report back to
// whoever submitted it. Field is a JSON path into Structured (experience[2].highlights[0]).
type Unsupported struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

// KeepSupported removes from s every value cvText does not contain, and reports each one.
//
// Strings and list items are checked by whole-word containment after normalisation. A role
// or project survives with whatever of it is supported: one with an unsupported company but
// supported bullets keeps the bullets, which the bank then holds as placeless evidence — an
// achievement is never lost because its context did not check out. Dates go through the
// Extractor's own year check (dropUnsupportedYears). TotalYears is an estimate no line of a
// CV prints verbatim, so it is kept only when the CV states that figure in years itself.
//
// It does not Sanitize; the caller does, as the Extractor does, after this.
func KeepSupported(s *Structured, cvText string) []Unsupported {
	src := newSourceText(cvText)
	var dropped []Unsupported

	keep := func(field string, v *string) {
		if strings.TrimSpace(*v) == "" || src.has(*v) {
			return
		}
		dropped = append(dropped, Unsupported{Field: field, Value: *v})
		*v = ""
	}
	keepAll := func(field string, vs []string) []string {
		out := vs[:0]
		for i, v := range vs {
			if strings.TrimSpace(v) == "" {
				continue
			}
			if src.has(v) {
				out = append(out, v)
				continue
			}
			dropped = append(dropped, Unsupported{Field: fmt.Sprintf("%s[%d]", field, i), Value: v})
		}
		return out
	}

	keep("full_name", &s.FullName)
	keep("headline", &s.Headline)
	keep("location", &s.Location)
	keep("email", &s.Email)
	keep("phone", &s.Phone)
	keep("summary", &s.Summary)
	s.Links = keepAll("links", s.Links)
	s.Languages = keepAll("languages", s.Languages)
	s.Skills = keepAll("skills", s.Skills)
	s.Certifications = keepAll("certifications", s.Certifications)

	if s.TotalYears != 0 && !src.has(fmt.Sprintf("%d years", s.TotalYears)) && !src.has(fmt.Sprintf("%d year", s.TotalYears)) {
		dropped = append(dropped, Unsupported{Field: "total_years", Value: fmt.Sprint(s.TotalYears)})
		s.TotalYears = 0
	}

	roles := s.Experience[:0]
	for i, e := range s.Experience {
		at := fmt.Sprintf("experience[%d]", i)
		keep(at+".title", &e.Title)
		keep(at+".company", &e.Company)
		keep(at+".location", &e.Location)
		keep(at+".summary", &e.Summary)
		e.Highlights = keepAll(at+".highlights", e.Highlights)
		e.Stack = keepAll(at+".stack", e.Stack)
		// "Present" is a claim that the role has not ended. A CV that never says so does not
		// support it, and a CV seeded from it would tell a recruiter the candidate still works there.
		if e.Current && !src.has("present") && !src.has("current") && !src.has("now") && !src.has("today") {
			dropped = append(dropped, Unsupported{Field: at + ".current", Value: "true"})
			e.Current = false
		}
		if e.Title == "" && e.Company == "" && len(e.Highlights) == 0 {
			continue
		}
		roles = append(roles, e)
	}
	s.Experience = roles

	schools := s.Education[:0]
	for i, e := range s.Education {
		at := fmt.Sprintf("education[%d]", i)
		keep(at+".degree", &e.Degree)
		keep(at+".institution", &e.Institution)
		if e.Degree == "" && e.Institution == "" {
			continue
		}
		schools = append(schools, e)
	}
	s.Education = schools

	projects := s.Projects[:0]
	for i, p := range s.Projects {
		at := fmt.Sprintf("projects[%d]", i)
		keep(at+".name", &p.Name)
		keep(at+".link", &p.Link)
		p.Highlights = keepAll(at+".highlights", p.Highlights)
		if p.Name == "" && len(p.Highlights) == 0 {
			continue
		}
		projects = append(projects, p)
	}
	s.Projects = projects

	if n := dropUnsupportedYears(s, cvText); n > 0 {
		dropped = append(dropped, Unsupported{Field: "dates", Value: fmt.Sprintf("%d date(s) whose year the CV never prints", n)})
	}
	return dropped
}
