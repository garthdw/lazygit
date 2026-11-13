package config

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

func TestCurrentDiffRendererName(t *testing.T) {
	tr := i18n.EnglishTranslationSet()

	scenarios := []struct {
		name               string
		diffRendererConfig DiffRendererConfig
		expected           string
	}{
		{
			name:               "explicit name takes precedence over the command",
			diffRendererConfig: DiffRendererConfig{Name: "delta side-by-side", Command: "delta --side-by-side"},
			expected:           "delta side-by-side",
		},
		{
			name:               "derived from the first word of the stdinFilter command",
			diffRendererConfig: DiffRendererConfig{Command: "delta --side-by-side"},
			expected:           "delta",
		},
		{
			name:               "surrounding whitespace in the command is ignored",
			diffRendererConfig: DiffRendererConfig{Command: "  diff-so-fancy  "},
			expected:           "diff-so-fancy",
		},
		{
			name:               "derived from the first word of the extDiff command",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --color=always"},
			expected:           "difft",
		},
		{
			name:               "no name can be derived for external diff",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff"},
			expected:           tr.ExternalDiffDiffRendererName,
		},
		{
			name:               "derived from first argument of rawGit args",
			diffRendererConfig: DiffRendererConfig{Type: "rawGit", Args: []string{"--color-words"}},
			expected:           "--color-words",
		},
		{
			name:               "no name can be derived for raw diff",
			diffRendererConfig: DiffRendererConfig{Type: "rawGit"},
			expected:           tr.DefaultDiffRendererName,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = []DiffRendererConfig{s.diffRendererConfig}
			config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			assert.Equal(t, s.expected, config.CurrentDiffRendererName(tr))
		})
	}
}

func TestGetStdinFilterCommand(t *testing.T) {
	scenarios := []struct {
		name               string
		diffRendererConfig DiffRendererConfig
		width              int
		lightBackground    bool
		expected           string
		expectedError      string
	}{
		{
			name:               "a command without template variables is passed through",
			diffRendererConfig: DiffRendererConfig{Command: "delta --paging=never"},
			width:              120,
			expected:           "delta --paging=never",
		},
		{
			name:               "the width the diff is rendered at",
			diffRendererConfig: DiffRendererConfig{Command: "delta --width={{width}}"},
			width:              120,
			expected:           "delta --width=120",
		},
		{
			name:               "the width of one side of a side-by-side rendering",
			diffRendererConfig: DiffRendererConfig{Command: "ydiff -p cat -w {{columnWidth}}"},
			width:              120,
			expected:           "ydiff -p cat -w 54",
		},
		{
			name:               "a template variable can also be written with a leading dot",
			diffRendererConfig: DiffRendererConfig{Command: "delta --width={{.width}}"},
			width:              120,
			expected:           "delta --width=120",
		},
		{
			name:               "the color scheme on a dark background",
			diffRendererConfig: DiffRendererConfig{Command: "delta --{{colorScheme}}"},
			width:              120,
			expected:           "delta --dark",
		},
		{
			name:               "the color scheme on a light background",
			diffRendererConfig: DiffRendererConfig{Command: "delta --{{colorScheme}}"},
			width:              120,
			lightBackground:    true,
			expected:           "delta --light",
		},
		{
			name:               "the command can choose between options by the color scheme",
			diffRendererConfig: DiffRendererConfig{Command: `delta --syntax-theme={{if eq .colorScheme "light"}}GitHub{{else}}Dracula{{end}}`},
			width:              120,
			lightBackground:    true,
			expected:           "delta --syntax-theme=GitHub",
		},
		{
			name:               "the command can use template expressions",
			diffRendererConfig: DiffRendererConfig{Command: "delta{{if gt .width 100}} --side-by-side{{end}}"},
			width:              120,
			expected:           "delta --side-by-side",
		},
		{
			name:               "an unknown template variable is an error",
			diffRendererConfig: DiffRendererConfig{Command: "delta --width={{.widht}}"},
			width:              120,
			expectedError:      "can't use the command 'delta --width={{.widht}}'",
		},
		{
			name:               "an unknown template variable without a leading dot is an error too",
			diffRendererConfig: DiffRendererConfig{Command: "delta --width={{widht}}"},
			width:              120,
			expectedError:      "can't use the command 'delta --width={{widht}}'",
		},
		{
			name:               "nothing is returned for a renderer of another type",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --width={{width}}"},
			width:              120,
			expected:           "",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = []DiffRendererConfig{s.diffRendererConfig}
			config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			command, err := config.GetStdinFilterCommand(DiffRendererValues{Width: s.width, LightBackground: s.lightBackground})
			if s.expectedError != "" {
				assert.ErrorContains(t, err, s.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, s.expected, command)
			}
		})
	}
}

func TestGetExternalDiffCommand(t *testing.T) {
	scenarios := []struct {
		name               string
		diffRendererConfig DiffRendererConfig
		expected           string
		expectedError      string
	}{
		{
			name:               "a command without template variables is passed through",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --color=always"},
			expected:           "difft --color=always",
		},
		{
			name:               "the width the diff is rendered at",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --width={{width}}"},
			expected:           "difft --width=120",
		},
		{
			name:               "the color scheme",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --background={{colorScheme}}"},
			expected:           "difft --background=dark",
		},
		{
			name:               "the width alongside the diff context size",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --width={{width}} --context={{diffContext}}"},
			expected:           "difft --width=120 --context=3",
		},
		{
			name:               "a variable of stdin filters is an error",
			diffRendererConfig: DiffRendererConfig{Type: "extDiff", Command: "difft --width={{columnWidth}}"},
			expectedError:      "can't use the command 'difft --width={{columnWidth}}'",
		},
		{
			name:               "nothing is returned for a renderer of another type",
			diffRendererConfig: DiffRendererConfig{Command: "delta --width={{width}}"},
			expected:           "",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = []DiffRendererConfig{s.diffRendererConfig}
			config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			command, err := config.GetExternalDiffCommand(DiffRendererValues{Width: 120, DiffContext: 3})
			if s.expectedError != "" {
				assert.ErrorContains(t, err, s.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, s.expected, command)
			}
		})
	}
}

func TestCurrentDiffRendererNameWithoutDiffRenderers(t *testing.T) {
	config := NewDiffRendererConfigManager(func() *UserConfig { return &UserConfig{} })

	tr := i18n.EnglishTranslationSet()
	assert.Equal(t, tr.DefaultDiffRendererName, config.CurrentDiffRendererName(tr))
}

func TestCycleDiffRenderers(t *testing.T) {
	userConfig := &UserConfig{}
	userConfig.Git.DiffRenderers = []DiffRendererConfig{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

	currentIndex := func() int {
		index, _ := config.CurrentDiffRendererIndex()
		return index
	}

	assert.Equal(t, 0, currentIndex())

	config.CycleDiffRenderers()
	assert.Equal(t, 1, currentIndex())
	config.CycleDiffRenderers()
	assert.Equal(t, 2, currentIndex())
	config.CycleDiffRenderers()
	assert.Equal(t, 0, currentIndex(), "cycling forward past the last diff renderer wraps to the first")

	config.CycleDiffRenderersBackward()
	assert.Equal(t, 2, currentIndex(), "cycling backward past the first diff renderer wraps to the last")
	config.CycleDiffRenderersBackward()
	assert.Equal(t, 1, currentIndex())
}

func TestFocusedDiffRenderer(t *testing.T) {
	diffRenderers := []DiffRendererConfig{
		{Type: "extDiff", Command: "difft --color=always"},
		{Name: "delta side-by-side", Command: "delta --side-by-side"},
		{Command: "delta --dark"},
		{Type: "rawGit", Args: []string{"--color-words"}},
		{Name: "twin", Command: "first"},
		{Name: "twin", Command: "second"},
	}

	scenarios := []struct {
		name                string
		focusedDiffRenderer string
		expectedOk          bool
		expectedIndex       int
	}{
		{
			name:                "none is configured",
			focusedDiffRenderer: "",
			expectedOk:          false,
		},
		{
			name:                "no entry has the name",
			focusedDiffRenderer: "nope",
			expectedOk:          false,
		},
		{
			name:                "matched by its explicit name",
			focusedDiffRenderer: "delta side-by-side",
			expectedOk:          true,
			expectedIndex:       1,
		},
		{
			name:                "an unnamed entry is matched by the first word of its command",
			focusedDiffRenderer: "delta",
			expectedOk:          true,
			expectedIndex:       2,
		},
		{
			name:                "a rawGit entry is matched by its first argument",
			focusedDiffRenderer: "--color-words",
			expectedOk:          true,
			expectedIndex:       3,
		},
		{
			name:                "the first of several entries with the name is used",
			focusedDiffRenderer: "twin",
			expectedOk:          true,
			expectedIndex:       4,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = diffRenderers
			userConfig.Git.FocusedDiffRenderer = s.focusedDiffRenderer
			config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			focused, ok := config.FocusedDiffRenderer()
			assert.Equal(t, s.expectedOk, ok)
			if ok {
				index, _ := focused.CurrentDiffRendererIndex()
				assert.Equal(t, s.expectedIndex, index)
			}
		})
	}
}

func TestFocusedDiffRendererIsPinnedToItsEntry(t *testing.T) {
	userConfig := &UserConfig{}
	userConfig.Git.DiffRenderers = []DiffRendererConfig{
		{Type: "extDiff", Command: "difft --color=always"},
		{Command: "delta --dark"},
	}
	userConfig.Git.FocusedDiffRenderer = "delta"
	config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

	focused, ok := config.FocusedDiffRenderer()
	assert.True(t, ok)

	values := DiffRendererValues{}
	assertFocusedIsDelta := func() {
		assert.Equal(t, DiffRendererType_StdinFilter, focused.GetDiffRendererType())
		command, err := focused.GetStdinFilterCommand(values)
		assert.NoError(t, err)
		assert.Equal(t, "delta --dark", command)
	}

	assertFocusedIsDelta()

	config.CycleDiffRenderers()
	config.CycleDiffRenderers()
	assertFocusedIsDelta()
	command, err := config.GetExternalDiffCommand(values)
	assert.NoError(t, err)
	assert.Equal(t, "difft --color=always", command, "the cycled renderer is unaffected")
}

func TestSwitchToFocusedDiffRenderer(t *testing.T) {
	userConfig := &UserConfig{}
	userConfig.Git.DiffRenderers = []DiffRendererConfig{
		{Name: "difft", Type: "extDiff", Command: "difft"},
		{Name: "delta", Command: "delta"},
		{Name: "git", Type: "rawGit"},
	}
	userConfig.Git.FocusedDiffRenderer = "delta"
	config := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })
	focused, _ := config.FocusedDiffRenderer()
	currentIndex := func() int {
		index, _ := config.CurrentDiffRendererIndex()
		return index
	}

	config.RestoreBrowsingDiffRenderer()
	assert.Equal(t, 0, currentIndex(), "nothing to restore without a switch")

	config.SwitchToFocusedDiffRenderer(focused)
	assert.Equal(t, 1, currentIndex())
	config.RestoreBrowsingDiffRenderer()
	assert.Equal(t, 0, currentIndex())

	config.SwitchToFocusedDiffRenderer(focused)
	config.CycleDiffRenderers()
	config.RestoreBrowsingDiffRenderer()
	assert.Equal(t, 2, currentIndex(), "a cycle after the switch is kept")

	config.CycleDiffRenderers()
	config.SwitchToFocusedDiffRenderer(focused)
	config.CycleDiffRenderersBackward()
	config.RestoreBrowsingDiffRenderer()
	assert.Equal(t, 0, currentIndex(), "a reverse cycle after the switch is kept")
}

func TestGetEntityPatchCommand(t *testing.T) {
	const command = `semdiff patch --entity {{entity}} {{if units}}--unit {{units}}{{end}} {{if staged}}--staged --reverse{{end}} -- {{path}} {{oldPath}}`
	quote := func(text string) string { return "'" + text + "'" }

	scenarios := []struct {
		name     string
		command  string
		values   EntityPatchValues
		expected string
		hasError bool
	}{
		{
			name:     "a whole entity is staged without units",
			command:  command,
			values:   EntityPatchValues{Entity: "src/a.ts::function::f", Path: "src/a.ts"},
			expected: `semdiff patch --entity 'src/a.ts::function::f'   -- 'src/a.ts' `,
		},
		{
			name:     "units and the unstaging flags",
			command:  command,
			values:   EntityPatchValues{Entity: "a.ts::function::f", Units: []int{1, 3}, Path: "a.ts", Unstage: true},
			expected: `semdiff patch --entity 'a.ts::function::f' --unit 1 3 --staged --reverse -- 'a.ts' `,
		},
		{
			name:     "a staged rename names both paths",
			command:  command,
			values:   EntityPatchValues{Entity: "new.ts::function::f", Path: "new.ts", OldPath: "old.ts", Unstage: true},
			expected: `semdiff patch --entity 'new.ts::function::f'  --staged --reverse -- 'new.ts' 'old.ts'`,
		},
		{
			name:     "the values can be written with a leading dot",
			command:  `p {{.entity}} {{.path}} {{.reverse}}`,
			values:   EntityPatchValues{Entity: "e", Path: "a b"},
			expected: `p 'e' 'a b' false`,
		},
		{
			name:     "an unknown variable is an error",
			command:  `p {{nope}}`,
			hasError: true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = []DiffRendererConfig{{Command: "semdiff --pager", EntityPatchCommand: s.command}}
			manager := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			assert.True(t, manager.HasEntityPatchCommand())
			result, err := manager.GetEntityPatchCommand(s.values, quote)
			if s.hasError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, s.expected, result)
		})
	}

	t.Run("a renderer without the command has none", func(t *testing.T) {
		userConfig := &UserConfig{}
		userConfig.Git.DiffRenderers = []DiffRendererConfig{{Command: "delta"}}
		manager := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

		assert.False(t, manager.HasEntityPatchCommand())
		result, err := manager.GetEntityPatchCommand(EntityPatchValues{}, quote)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("no renderers have none either", func(t *testing.T) {
		manager := NewDiffRendererConfigManager(func() *UserConfig { return &UserConfig{} })

		assert.False(t, manager.HasEntityPatchCommand())
	})
}

func TestGetConsumersCommand(t *testing.T) {
	const command = `semdiff impact --entity {{entity}} --depth {{depth}} --json -- {{path}}`
	quote := func(text string) string { return "'" + text + "'" }

	scenarios := []struct {
		name     string
		command  string
		values   ConsumersValues
		expected string
		hasError bool
	}{
		{
			name:     "the text values are quoted and the depth is not",
			command:  command,
			values:   ConsumersValues{Entity: "src/a.ts::function::f;1", Depth: 2, Path: "src/a b.ts"},
			expected: `semdiff impact --entity 'src/a.ts::function::f;1' --depth 2 --json -- 'src/a b.ts'`,
		},
		{
			name:     "no depth means all of them",
			command:  command,
			values:   ConsumersValues{Entity: "e", Path: "p"},
			expected: `semdiff impact --entity 'e' --depth 0 --json -- 'p'`,
		},
		{
			name:     "the values can be written with a leading dot",
			command:  `c {{.entity}} {{.depth}}`,
			values:   ConsumersValues{Entity: "e", Depth: 1},
			expected: `c 'e' 1`,
		},
		{
			name:     "an unknown variable is an error",
			command:  `c {{nope}}`,
			hasError: true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			userConfig := &UserConfig{}
			userConfig.Git.DiffRenderers = []DiffRendererConfig{{Command: "semdiff --pager", ConsumersCommand: s.command}}
			manager := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

			assert.True(t, manager.HasConsumersCommand())
			result, err := manager.GetConsumersCommand(s.values, quote)
			if s.hasError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, s.expected, result)
		})
	}

	t.Run("a renderer without the command has none", func(t *testing.T) {
		userConfig := &UserConfig{}
		userConfig.Git.DiffRenderers = []DiffRendererConfig{{Command: "delta", EntityPatchCommand: "x"}}
		manager := NewDiffRendererConfigManager(func() *UserConfig { return userConfig })

		assert.False(t, manager.HasConsumersCommand())
		result, err := manager.GetConsumersCommand(ConsumersValues{}, quote)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}
