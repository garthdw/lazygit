package controllers

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestEntityTargets(t *testing.T) {
	row := func(entity string, unit int, lineType types.DiffLineType) types.DiffLineInfo {
		return types.DiffLineInfo{Path: "/repo/a.ts", EntityID: entity, Unit: unit, Type: lineType}
	}
	const (
		add = "a.ts::function::add"
		sub = "a.ts::function::sub"
	)

	scenarios := []struct {
		name     string
		infos    []types.DiffLineInfo
		byUnit   bool
		expected []entityTarget
	}{
		{
			name:     "rows that say nothing about an entity are left out",
			infos:    []types.DiffLineInfo{{Path: "/repo/a.ts", Type: types.DiffLineAdded}},
			expected: []entityTarget{},
		},
		{
			name: "an entity once, however many rows of it are selected",
			infos: []types.DiffLineInfo{
				row(add, 0, types.DiffLineSpan), row(add, 1, types.DiffLineDeleted), row(add, 2, types.DiffLineAdded),
				row(sub, 0, types.DiffLineSpan),
			},
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add}, {path: "/repo/a.ts", entityID: sub}},
		},
		{
			name:     "a unit row stands for its unit",
			infos:    []types.DiffLineInfo{row(add, 2, types.DiffLineAdded)},
			byUnit:   true,
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add, units: []int{2}}},
		},
		{
			name: "units are collected, once each and in order",
			infos: []types.DiffLineInfo{
				row(add, 3, types.DiffLineAdded), row(add, 1, types.DiffLineDeleted), row(add, 1, types.DiffLineAdded),
			},
			byUnit:   true,
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add, units: []int{1, 3}}},
		},
		{
			name:     "a header stands for the whole entity",
			infos:    []types.DiffLineInfo{row(add, 0, types.DiffLineSpan)},
			byUnit:   true,
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add}},
		},
		{
			name:     "a row of an added entity, which has no units, stands for the whole entity",
			infos:    []types.DiffLineInfo{row(add, 0, types.DiffLineAdded)},
			byUnit:   true,
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add}},
		},
		{
			name:     "a header selected along with units adds nothing to them",
			infos:    []types.DiffLineInfo{row(add, 0, types.DiffLineSpan), row(add, 2, types.DiffLineAdded)},
			byUnit:   true,
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add, units: []int{2}}},
		},
		{
			name:     "context between units stands for nothing",
			infos:    []types.DiffLineInfo{row(add, 0, types.DiffLineContext)},
			byUnit:   true,
			expected: []entityTarget{},
		},
		{
			name:     "context between units is still part of the entity",
			infos:    []types.DiffLineInfo{row(add, 0, types.DiffLineContext)},
			expected: []entityTarget{{path: "/repo/a.ts", entityID: add}},
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			assert.Equal(t, s.expected, entityTargets(s.infos, s.byUnit))
		})
	}
}
