package resumeextract

import (
	"slices"
	"testing"

	"github.com/strelov1/freehire/internal/candidate/perioddate"
)

const verifyCV = `Jane Doe
Senior Backend Engineer · Berlin
jane@example.com

SUMMARY
Backend engineer building payment systems in Go.

EXPERIENCE
Acme Pay — Senior Backend Engineer          03/2021 – Present
• Cut checkout latency from 900ms to 120ms by moving fraud
  scoring off the request path.
• Led a team of five engineers.

Globex — Software Engineer                  2017 – 2021
• Built the billing pipeline in Python.

EDUCATION
TU Berlin, BSc Computer Science, 2017

Languages: English, German
Certifications: AWS Certified Developer
Skills: Go, PostgreSQL, Kafka
8 years of experience`

func fields(us []Unsupported) []string {
	out := make([]string, 0, len(us))
	for _, u := range us {
		out = append(out, u.Field)
	}
	return out
}

func TestKeepSupported_KeepsWhatTheCVSaysAndDropsWhatItDoesNot(t *testing.T) {
	s := Structured{
		FullName: "Jane Doe",
		Headline: "Senior Backend Engineer",
		Summary:  "Backend engineer building payment systems in Go.",
		Experience: []Experience{
			{
				Title:   "Senior Backend Engineer",
				Company: "Acme Pay",
				Start:   &perioddate.PeriodDate{Year: 2021, Month: 3},
				Current: true,
				Highlights: []string{
					// Wrapped across two lines in the PDF: whitespace must not matter.
					"Cut checkout latency from 900ms to 120ms by moving fraud scoring off the request path.",
					"Led a team of five engineers",
					// Invented: the CV says five, not twelve.
					"Led a team of twelve engineers",
				},
				Stack: []string{"Go", "Rust"},
			},
		},
		Education:      []Education{{Degree: "BSc Computer Science", Institution: "TU Berlin"}},
		Languages:      []string{"English", "German", "French"},
		Certifications: []string{"AWS Certified Developer", "CKA"},
		Skills:         []string{"Go", "PostgreSQL"},
		TotalYears:     8,
	}

	dropped := KeepSupported(&s, verifyCV)

	want := []string{
		"languages[2]", "certifications[1]",
		"experience[0].highlights[2]", "experience[0].stack[1]",
	}
	if got := fields(dropped); !slices.Equal(got, want) {
		t.Fatalf("dropped = %v, want %v", got, want)
	}
	if len(s.Experience) != 1 || len(s.Experience[0].Highlights) != 2 {
		t.Fatalf("experience = %+v, want one role with the two supported bullets", s.Experience)
	}
	if !s.Experience[0].Current {
		t.Error("Current dropped although the CV says Present")
	}
	if s.TotalYears != 8 {
		t.Errorf("TotalYears = %d, want 8 (the CV says 8 years)", s.TotalYears)
	}
}

// A role whose title the CV never gave keeps its supported bullets: losing the evidence
// because its context failed to check out would be the worse error.
func TestKeepSupported_AnInventedTitleDoesNotCostTheBullets(t *testing.T) {
	s := Structured{Experience: []Experience{{
		Title:      "Staff Engineer",
		Company:    "Globex",
		Highlights: []string{"Built the billing pipeline in Python."},
	}}}

	dropped := KeepSupported(&s, verifyCV)

	if got := fields(dropped); !slices.Equal(got, []string{"experience[0].title"}) {
		t.Fatalf("dropped = %v", got)
	}
	if e := s.Experience[0]; e.Title != "" || e.Company != "Globex" || len(e.Highlights) != 1 {
		t.Fatalf("role = %+v", e)
	}
}

func TestKeepSupported_ARoleWithNothingSupportedIsRemoved(t *testing.T) {
	s := Structured{Experience: []Experience{{Title: "CTO", Company: "Initech", Highlights: []string{"Raised $10M."}}}}

	KeepSupported(&s, verifyCV)

	if len(s.Experience) != 0 {
		t.Fatalf("experience = %+v, want none", s.Experience)
	}
}

// Containment is by whole words: a short value must not be found inside a longer one.
func TestKeepSupported_MatchesWholeWordsOnly(t *testing.T) {
	s := Structured{Skills: []string{"Java", "Ka"}, TotalYears: 7}

	KeepSupported(&s, "Skills: JavaScript, Kafka. 2017 – 2021")

	if len(s.Skills) != 0 {
		t.Errorf("skills = %v, want none (JavaScript is not Java)", s.Skills)
	}
	if s.TotalYears != 0 {
		t.Errorf("TotalYears = %d, want 0 (the CV states no such figure)", s.TotalYears)
	}
}

// PDF text layers carry ligatures that print as ordinary letters.
func TestKeepSupported_FoldsPresentationForms(t *testing.T) {
	s := Structured{Experience: []Experience{{Highlights: []string{"Shipped the first offline mode."}}}}

	KeepSupported(&s, "• Shipped the ﬁrst oﬄine mode.")

	if len(s.Experience) != 1 || len(s.Experience[0].Highlights) != 1 {
		t.Fatalf("experience = %+v, want the ligature-spelled bullet kept", s.Experience)
	}
}

func TestKeepSupported_PresentNeedsTheCVToSaySo(t *testing.T) {
	s := Structured{Experience: []Experience{{Company: "Globex", Current: true}}}

	dropped := KeepSupported(&s, "Globex — Software Engineer 2017 – 2021")

	if s.Experience[0].Current {
		t.Error("Current kept although the CV gives an end year")
	}
	if got := fields(dropped); !slices.Equal(got, []string{"experience[0].current"}) {
		t.Fatalf("dropped = %v", got)
	}
}

func TestKeepSupported_DropsAYearTheCVNeverPrints(t *testing.T) {
	s := Structured{Experience: []Experience{{
		Company: "Globex",
		Start:   &perioddate.PeriodDate{Year: 2015},
		End:     &perioddate.PeriodDate{Year: 2021},
	}}}

	KeepSupported(&s, verifyCV)

	if e := s.Experience[0]; e.Start != nil || e.End == nil {
		t.Fatalf("dates = %v / %v, want the unsupported start dropped and the end kept", e.Start, e.End)
	}
}
