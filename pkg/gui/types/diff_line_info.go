package types

import "github.com/jesseduffield/lazygit/pkg/commands/patch"

// DiffLineType classifies a row of a rendered diff.
type DiffLineType int

const (
	DiffLineFileHeader DiffLineType = iota
	DiffLineHunkHeader
	DiffLineContext
	DiffLineAdded
	DiffLineDeleted
	// DiffLineOther is anything that isn't one of the above, e.g. the
	// "\ No newline at end of file" marker.
	DiffLineOther
	// DiffLineSpan is a row standing for a group of change lines, such as a whole
	// function, as the ranges NewLine–NewEnd (additions) and OldLine–OldEnd
	// (deletions) state.
	DiffLineSpan
)

// DiffLineInfo is the identity of a row of a rendered diff in terms of the patch
// it was rendered from: which file the row belongs to, what kind of row it is,
// and where the line sits in the old and new versions of that file. This lets us
// act on the line the user is pointing at in a diff view: stage it, open it in an
// editor, keep the cursor on it across a re-render. The rendered text alone tells
// us none of that.
type DiffLineInfo struct {
	// Path is the absolute path of the file the line belongs to.
	Path string
	Type DiffLineType
	// NewLine is the line's position in the new version of the file. Set for all
	// content lines (for a deletion it is the position the deletion sits at) and
	// for hunk headers (the first line of the hunk they head).
	NewLine int
	// OldLine is the line's position in the old version of the file. Set only
	// for deletions, which are the only rows that need it: two consecutive
	// deletions share a new-file position and differ only here.
	OldLine int
	// NewEnd and OldEnd end a span's ranges; 0 where the span has no such side.
	NewEnd int
	OldEnd int
	// EntityID is the id of the entity (a function, a class) the renderer says the row
	// belongs to, empty when it says nothing. Unit is the 1-based change unit of that
	// entity the row is in, 0 for a row outside every unit, such as the entity's header.
	EntityID string
	Unit     int
}

// IsChange reports whether the row is an added or deleted line, as opposed to a
// context line or a header. It mirrors patch.PatchLine.IsChange: those are the
// rows a patch is built from, and the rows navigation moves between.
func (self DiffLineInfo) IsChange() bool {
	return self.Type == DiffLineAdded || self.Type == DiffLineDeleted
}

// PatchLineIdentity says which change line of the file the row is, in the terms a patch
// of that file is built and read in: an addition by where it sits in the new version of
// the file, a deletion by where it sat in the old one. Only meaningful for a change row.
func (self DiffLineInfo) PatchLineIdentity() patch.LineIdentity {
	if self.Type == DiffLineDeleted {
		return patch.LineIdentity{LineNumber: self.OldLine, IsDeletion: true}
	}
	return patch.LineIdentity{LineNumber: self.NewLine}
}

// PatchLineIdentities says which change lines of the file the row stands for: a change
// row's own, or a span's additions and deletions.
func (self DiffLineInfo) PatchLineIdentities() []patch.LineIdentity {
	if self.Type != DiffLineSpan {
		return []patch.LineIdentity{self.PatchLineIdentity()}
	}
	identities := []patch.LineIdentity{}
	if self.NewEnd > 0 {
		identities = append(identities, patch.LineIdentity{LineNumber: self.NewLine, LastLineNumber: self.NewEnd})
	}
	if self.OldEnd > 0 {
		identities = append(identities,
			patch.LineIdentity{LineNumber: self.OldLine, LastLineNumber: self.OldEnd, IsDeletion: true})
	}
	return identities
}

// IsContent reports whether the row is a line of the file itself — a change or a
// context line — as opposed to a header or a marker. Those are the rows that have a
// position in the file, and so can be looked for in another rendering of the same
// diff, or in the diff itself.
func (self DiffLineInfo) IsContent() bool {
	return self.IsChange() || self.Type == DiffLineContext
}
