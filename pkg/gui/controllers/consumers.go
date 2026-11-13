package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
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

	rows := make([]consumerRow, 0, len(dependents)+len(self.Tests))
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
	rows := make([]consumerRow, 0, len(self.PossibleCallers))
	for _, caller := range self.PossibleCallers {
		line := caller.Line
		if len(caller.Sites) == 1 {
			line = caller.Sites[0].Line
		}
		sites := make([]string, 0, len(caller.Sites))
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
	reasons := make([]string, 0, len(self.IncompleteReasons))
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

// showConsumersDescription lists the key only under a renderer that can be asked for
// consumers, so that it stays out of the way of everyone who has none; pressing the key
// still says why it does nothing.
func (self *MainViewController) showConsumersDescription() string {
	if !self.c.State().GetDiffRendererConfigManager().HasConsumersCommand() {
		return ""
	}
	return self.c.Tr.ShowConsumers
}

func (self *MainViewController) showConsumersDisabledReason() *types.DisabledReason {
	if reason := self.diffSelectionDisabledReason(); reason != nil {
		return reason
	}
	if !self.c.State().GetDiffRendererConfigManager().HasConsumersCommand() {
		return &types.DisabledReason{Text: self.c.Tr.NoConsumersCommand}
	}
	return nil
}

// showConsumers opens the popup of what depends on the entity of the selected row, which
// the diff renderer's own records name.
func (self *MainViewController) showConsumers() error {
	view := self.context.GetView()
	selected := view.SelectedLineIdx()
	infos := self.c.Helpers().DiffLine.DiffLinesInBufferRange(view, selected, selected)
	info, found := lo.Find(infos, func(info types.DiffLineInfo) bool { return info.EntityID != "" })
	if !found {
		return errors.New(self.c.Tr.NoEntityAtSelection)
	}

	values := config.ConsumersValues{
		Entity: info.EntityID,
		Depth:  1,
		Path:   repoRelativePath(self.c.Git().RepoPaths.WorktreePath(), info.Path),
	}
	return self.openConsumers(values)
}

// openConsumers runs the consumersCommand away from the UI thread, as it may take a
// moment, and shows what it found. Changing the depth opens it again.
func (self *MainViewController) openConsumers(values config.ConsumersValues) error {
	command, err := self.c.State().GetDiffRendererConfigManager().GetConsumersCommand(values, self.c.OS().Quote)
	if err != nil {
		return err
	}
	if command == "" {
		return errors.New(self.c.Tr.NoConsumersCommand)
	}

	return self.c.WithWaitingStatus(self.c.Tr.FindingConsumers, func(gocui.Task) error {
		stdout, stderr, err := self.c.OS().Cmd.NewShell(command, "").RunWithOutputs()
		if err != nil {
			reason, _, _ := strings.Cut(strings.TrimSpace(stderr), "\n")
			if reason == "" {
				reason = err.Error()
			}
			return errors.New(utils.ResolvePlaceholderString(self.c.Tr.ConsumersCommandFailed,
				map[string]string{"error": reason}))
		}
		report, err := parseConsumers(stdout)
		if err != nil {
			return errors.New(utils.ResolvePlaceholderString(self.c.Tr.ConsumersCommandBadOutput,
				map[string]string{"error": err.Error()}))
		}

		self.c.OnUIThread(func() error {
			return self.c.Menu(self.consumersMenu(report, values))
		})
		return nil
	})
}

func (self *MainViewController) consumersMenu(report consumersReport, values config.ConsumersValues) types.CreateMenuOptions {
	depth := func(depth int) string {
		if depth == 0 {
			return self.c.Tr.ConsumersDepthAll
		}
		return fmt.Sprint(depth)
	}

	jumpTo := func(row consumerRow) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: []string{row.Kind, row.Name, row.location()},
			Tooltip:      row.Detail,
			OnPress: func() error {
				path := filepath.Join(self.c.Git().RepoPaths.WorktreePath(), filepath.FromSlash(row.File))
				return self.c.Helpers().Files.EditFileAtLine(path, row.Line)
			},
		}
	}

	items := []*types.MenuItem{}
	for _, row := range report.confirmed() {
		items = append(items, jumpTo(row))
	}
	if len(items) == 0 {
		items = append(items, &types.MenuItem{
			Label:   self.c.Tr.NoConsumersFound,
			OnPress: func() error { return nil },
		})
	}

	var prompt string
	if report.Basis == "textual" {
		prompt = utils.ResolvePlaceholderString(self.c.Tr.ConsumersApproximate,
			map[string]string{"name": report.Entity.Name})
	}
	if !report.Complete {
		possible := report.possible()
		reasons := map[string]string{"count": fmt.Sprint(len(possible)), "reasons": report.incompleteReasons()}
		if len(possible) == 0 {
			incomplete := utils.ResolvePlaceholderString(self.c.Tr.ConsumersIncompleteNoCallers, reasons)
			prompt = strings.TrimSpace(prompt + "\n" + incomplete)
		}
		section := &types.MenuSection{Title: utils.ResolvePlaceholderString(self.c.Tr.ConsumersIncomplete, reasons)}
		for _, row := range possible {
			item := jumpTo(row)
			item.Section = section
			items = append(items, item)
		}
	}

	next := nextConsumersDepth(values.Depth)
	// In the last column, so that it doesn't widen the ones the rows are aligned in.
	items = append(items, &types.MenuItem{
		LabelColumns: []string{"", "", utils.ResolvePlaceholderString(self.c.Tr.ConsumersChangeDepth,
			map[string]string{"depth": depth(values.Depth), "next": depth(next)})},
		Keys: []gocui.Key{gocui.NewKeyRune('d')},
		OnPress: func() error {
			values.Depth = next
			return self.openConsumers(values)
		},
	})

	return types.CreateMenuOptions{
		Title: utils.ResolvePlaceholderString(self.c.Tr.ConsumersTitle,
			map[string]string{"entity": report.entityName(), "depth": depth(values.Depth)}),
		Prompt:     prompt,
		Items:      items,
		HideCancel: true,
	}
}
