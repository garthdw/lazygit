package controllers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// consumersReport is the JSON that a diff renderer's consumersCommand prints. Fields
// it has that are not listed here are left alone.
type consumersReport struct {
	Entity struct {
		Name       string   `json:"name"`
		Breadcrumb []string `json:"breadcrumb"`
	} `json:"entity"`
	Basis             string `json:"basis"`
	Complete          bool   `json:"complete"`
	IncompleteReasons []struct {
		Detail string `json:"detail"`
	} `json:"incompleteReasons"`
	Dependents      []consumersLocation `json:"dependents"`
	Tests           []consumersLocation `json:"tests"`
	PossibleCallers []struct {
		Name  string `json:"name"`
		File  string `json:"file"`
		Line  int    `json:"line"`
		Sites []struct {
			Kind string `json:"kind"`
			Line int    `json:"line"`
		} `json:"sites"`
	} `json:"possibleCallers"`
}

type consumersLocation struct {
	Name       string   `json:"name"`
	Breadcrumb []string `json:"breadcrumb"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Depth      int      `json:"depth"`
}

// consumerRow is a line of the consumers popup: what kind of consumer it is (`d1`, `test`
// or `?`), what it is called, and where to go to see it.
type consumerRow struct {
	Kind   string
	Name   string
	File   string
	Line   int
	Detail string
}

func (self consumerRow) location() string {
	return fmt.Sprintf("%s:%d", self.File, self.Line)
}

func parseConsumers(output string) (consumersReport, error) {
	var report consumersReport
	err := json.Unmarshal([]byte(output), &report)
	return report, err
}

// breadcrumbWithoutFile names an entity by what contains it; the first element of a
// breadcrumb is its file, which the rows give on their own.
func breadcrumbWithoutFile(breadcrumb []string, fallback string) string {
	if len(breadcrumb) > 1 {
		breadcrumb = breadcrumb[1:]
	}
	if len(breadcrumb) == 0 {
		return fallback
	}
	return strings.Join(breadcrumb, " › ")
}

func (self consumersReport) entityName() string {
	return breadcrumbWithoutFile(self.Entity.Breadcrumb, self.Entity.Name)
}

// confirmed returns the dependents, nearest first, then the tests.
func (self consumersReport) confirmed() []consumerRow {
	dependents := slices.Clone(self.Dependents)
	slices.SortStableFunc(dependents, func(a, b consumersLocation) int { return a.Depth - b.Depth })

	rows := []consumerRow{}
	for _, dependent := range dependents {
		rows = append(rows, consumerRow{
			Kind: fmt.Sprintf("d%d", dependent.Depth),
			Name: breadcrumbWithoutFile(dependent.Breadcrumb, dependent.Name),
			File: dependent.File,
			Line: dependent.Line,
		})
	}
	for _, test := range self.Tests {
		rows = append(rows, consumerRow{
			Kind: "test",
			Name: breadcrumbWithoutFile(test.Breadcrumb, test.Name),
			File: test.File,
			Line: test.Line,
		})
	}
	return rows
}

// possible returns the things that may call the entity without the graph showing it.
// A caller with a single place that refers to the entity is opened at that place rather
// than at the caller's own start.
func (self consumersReport) possible() []consumerRow {
	rows := []consumerRow{}
	for _, caller := range self.PossibleCallers {
		line := caller.Line
		if len(caller.Sites) == 1 {
			line = caller.Sites[0].Line
		}
		sites := []string{}
		for _, site := range caller.Sites {
			sites = append(sites, fmt.Sprintf("%s at line %d", site.Kind, site.Line))
		}
		rows = append(rows, consumerRow{
			Kind:   "?",
			Name:   caller.Name,
			File:   caller.File,
			Line:   line,
			Detail: strings.Join(sites, "\n"),
		})
	}
	return rows
}

func (self consumersReport) incompleteReasons() string {
	reasons := []string{}
	for _, reason := range self.IncompleteReasons {
		reasons = append(reasons, reason.Detail)
	}
	return strings.Join(reasons, "; ")
}

// nextConsumersDepth cycles the depth of the dependents listed: 1, 2, all (0), and 1 again.
func nextConsumersDepth(depth int) int {
	switch depth {
	case 1:
		return 2
	case 2:
		return 0
	default:
		return 1
	}
}
