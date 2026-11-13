package config

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/jesseduffield/lazygit/pkg/i18n"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

type DiffRendererConfigManager struct {
	getUserConfig     func() *UserConfig
	diffRendererIndex int
}

type DiffRendererType int

const (
	DiffRendererType_StdinFilter DiffRendererType = iota
	DiffRendererType_ExtDiff
	DiffRendererType_RawGit
)

func NewDiffRendererConfigManager(getUserConfig func() *UserConfig) *DiffRendererConfigManager {
	return &DiffRendererConfigManager{getUserConfig: getUserConfig}
}

func (self *DiffRendererConfigManager) currentDiffRendererConfig() *DiffRendererConfig {
	diffRenderers := self.getUserConfig().Git.DiffRenderers
	if len(diffRenderers) == 0 {
		return nil
	}

	// Guard against the diff renderer index being out of range, which can happen if the user
	// has removed diff renderers from their config file while lazygit is running.
	if self.diffRendererIndex >= len(diffRenderers) {
		self.diffRendererIndex = 0
	}

	return &diffRenderers[self.diffRendererIndex]
}

func (self *DiffRendererConfig) getType() DiffRendererType {
	switch self.Type {
	case "stdinFilter", "":
		return DiffRendererType_StdinFilter
	case "extDiff":
		return DiffRendererType_ExtDiff
	case "rawGit":
		return DiffRendererType_RawGit
	}
	panic("invalid diff renderer type: " + self.Type)
}

func (self *DiffRendererConfigManager) GetDiffRendererType() DiffRendererType {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil {
		return DiffRendererType_RawGit
	}
	return currentDiffRendererConfig.getType()
}

// DiffRendererValues are what the command of a diff renderer can refer to.
type DiffRendererValues struct {
	// The width of the view that the diff is rendered into
	Width int
	// The number of lines of context around each hunk
	DiffContext uint64
	// Whether the terminal has a light background
	LightBackground bool
}

func (self *DiffRendererConfigManager) GetStdinFilterCommand(values DiffRendererValues) (string, error) {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil || currentDiffRendererConfig.getType() != DiffRendererType_StdinFilter {
		return "", nil
	}

	return currentDiffRendererConfig.resolveCommand(values)
}

func (self *DiffRendererConfigManager) GetColorArg() string {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil || currentDiffRendererConfig.getType() != DiffRendererType_StdinFilter {
		return "always"
	}

	colorArg := currentDiffRendererConfig.ColorArg
	if colorArg == "" {
		return "always"
	}
	return colorArg
}

func (self *DiffRendererConfigManager) GetExternalDiffCommand(values DiffRendererValues) (string, error) {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil || currentDiffRendererConfig.getType() != DiffRendererType_ExtDiff {
		return "", nil
	}

	return currentDiffRendererConfig.resolveCommand(values)
}

// resolveCommand resolves the renderer's command, which is a Go template with
// the values it can refer to as its variables. A variable can be written with
// or without the leading dot, as in {{.width}} or {{width}}.
func (self *DiffRendererConfig) resolveCommand(values DiffRendererValues) (string, error) {
	colorScheme := "dark"
	if values.LightBackground {
		colorScheme = "light"
	}
	variables := map[string]any{
		"width":       values.Width,
		"colorScheme": colorScheme,
	}
	switch self.getType() {
	case DiffRendererType_StdinFilter:
		variables["columnWidth"] = values.Width/2 - 6
	case DiffRendererType_ExtDiff:
		variables["diffContext"] = values.DiffContext
	case DiffRendererType_RawGit:
		// has no command
	}

	return resolveRendererTemplate("command", string(self.Command), variables)
}

// resolveRendererTemplate resolves a Go template of a diff renderer, in which each
// variable can be written with or without the leading dot.
func resolveRendererTemplate(what string, text string, variables map[string]any) (string, error) {
	funcs := template.FuncMap{}
	for name, value := range variables {
		funcs[name] = func() any { return value }
	}

	resolved, err := utils.ResolveTemplate(text, variables, funcs)
	if err != nil {
		return "", fmt.Errorf("git.diffRenderers: can't use the %s '%s': %w", what, text, err)
	}
	return resolved, nil
}

// EntityPatchValues are what the entityPatchCommand of a diff renderer can refer to.
type EntityPatchValues struct {
	Entity string
	// The change units to stage; none stages the whole entity.
	Units []int
	Path  string
	// The path the file had before a rename, when unstaging a staged one.
	OldPath string
	// Whether the patch takes the entity out of the index, rather than putting it in.
	Unstage bool
}

// HasEntityPatchCommand reports whether the current renderer can be asked for the
// patch of an entity.
func (self *DiffRendererConfigManager) HasEntityPatchCommand() bool {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	return currentDiffRendererConfig != nil && strings.TrimSpace(currentDiffRendererConfig.EntityPatchCommand) != ""
}

// GetEntityPatchCommand resolves the current renderer's entityPatchCommand, a Go
// template like the renderer's command. The text values are passed through quote
// before they are substituted, because the result is run by a shell and an entity id
// or a path is nobody's idea of a command.
func (self *DiffRendererConfigManager) GetEntityPatchCommand(values EntityPatchValues, quote func(string) string) (string, error) {
	if !self.HasEntityPatchCommand() {
		return "", nil
	}
	currentDiffRendererConfig := self.currentDiffRendererConfig()

	quoteUnlessEmpty := func(text string) string {
		if text == "" {
			return ""
		}
		return quote(text)
	}
	units := strings.Join(lo.Map(values.Units, func(unit int, _ int) string { return strconv.Itoa(unit) }), " ")
	variables := map[string]any{
		"entity":  quote(values.Entity),
		"path":    quote(values.Path),
		"oldPath": quoteUnlessEmpty(values.OldPath),
		"units":   units,
		"staged":  values.Unstage,
		"reverse": values.Unstage,
	}
	return resolveRendererTemplate("entityPatchCommand", currentDiffRendererConfig.EntityPatchCommand, variables)
}

// ConsumersValues are what the consumersCommand of a diff renderer can refer to.
type ConsumersValues struct {
	Entity string
	// How many steps of dependents to list; 0 lists all of them.
	Depth int
	Path  string
}

// HasConsumersCommand reports whether the current renderer can be asked for the
// consumers of an entity.
func (self *DiffRendererConfigManager) HasConsumersCommand() bool {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	return currentDiffRendererConfig != nil && strings.TrimSpace(currentDiffRendererConfig.ConsumersCommand) != ""
}

// GetConsumersCommand resolves the current renderer's consumersCommand, a Go template
// like the entityPatchCommand, whose text values are quoted the same way.
func (self *DiffRendererConfigManager) GetConsumersCommand(values ConsumersValues, quote func(string) string) (string, error) {
	if !self.HasConsumersCommand() {
		return "", nil
	}
	currentDiffRendererConfig := self.currentDiffRendererConfig()

	variables := map[string]any{
		"entity": quote(values.Entity),
		"depth":  strconv.Itoa(values.Depth),
		"path":   quote(values.Path),
	}
	return resolveRendererTemplate("consumersCommand", currentDiffRendererConfig.ConsumersCommand, variables)
}

func (self *DiffRendererConfigManager) GetRawGitArgs() []string {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil || currentDiffRendererConfig.getType() != DiffRendererType_RawGit {
		return nil
	}
	return currentDiffRendererConfig.Args
}

func (self *DiffRendererConfigManager) CycleDiffRenderers() {
	self.diffRendererIndex = (self.diffRendererIndex + 1) % len(self.getUserConfig().Git.DiffRenderers)
}

func (self *DiffRendererConfigManager) CycleDiffRenderersBackward() {
	n := len(self.getUserConfig().Git.DiffRenderers)
	self.diffRendererIndex = (self.diffRendererIndex - 1 + n) % n
}

func (self *DiffRendererConfigManager) CurrentDiffRendererIndex() (int, int) {
	return self.diffRendererIndex, len(self.getUserConfig().Git.DiffRenderers)
}

// CurrentDiffRendererName returns a name for the current diff renderer, suitable for showing
// to the user.
func (self *DiffRendererConfigManager) CurrentDiffRendererName(tr *i18n.TranslationSet) string {
	currentDiffRendererConfig := self.currentDiffRendererConfig()
	if currentDiffRendererConfig == nil {
		return tr.DefaultDiffRendererName
	}

	if name := currentDiffRendererConfig.displayName(); name != "" {
		return name
	}

	if currentDiffRendererConfig.getType() == DiffRendererType_ExtDiff && currentDiffRendererConfig.Command == "" {
		return tr.ExternalDiffDiffRendererName
	}

	return tr.DefaultDiffRendererName
}

func (self *DiffRendererConfig) displayName() string {
	if self.Name != "" {
		return self.Name
	}
	if self.getType() == DiffRendererType_RawGit && len(self.Args) > 0 {
		return self.Args[0]
	}
	return firstWord(string(self.Command))
}

func firstWord(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
