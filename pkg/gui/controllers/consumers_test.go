package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const consumersJSON = `{
  "schemaVersion": 1,
  "somethingNew": {"ignored": true},
  "entity": {"id": "src/cart.ts::function::formatPrice", "name": "formatPrice", "kind": "function",
             "breadcrumb": ["src/cart.ts", "formatPrice"], "file": "src/cart.ts"},
  "basis": "graph", "depth": 2, "complete": false,
  "incompleteReasons": [{"code": "value_refs", "detail": "1 name used as a value"}, {"code": "x", "detail": "another"}],
  "dependents": [
    {"name": "render", "breadcrumb": ["src/checkout.ts", "Page", "render"], "file": "src/checkout.ts", "line": 12, "depth": 2},
    {"name": "onAdd", "breadcrumb": ["src/checkout.ts", "onAdd"], "file": "src/checkout.ts", "line": 3, "depth": 1}
  ],
  "tests": [{"name": "formats", "breadcrumb": ["src/cart.spec.ts", "cart", "formats"], "file": "src/cart.spec.ts", "line": 5}],
  "possibleCallers": [
    {"name": "register", "kind": "function", "file": "src/events.ts", "line": 8, "sites": [{"kind": "value_ref", "line": 9}]},
    {"name": "wire", "kind": "function", "file": "src/events.ts", "line": 20, "sites": [{"kind": "value_ref", "line": 21}, {"kind": "value_ref", "line": 25}]}
  ]
}`

func TestParseConsumers(t *testing.T) {
	report, err := parseConsumers(consumersJSON)
	assert.NoError(t, err)

	assert.Equal(t, "formatPrice", report.entityName())
	assert.False(t, report.Complete)
	assert.Equal(t, "1 name used as a value; another", report.incompleteReasons())

	assert.Equal(t, []consumerRow{
		{Kind: "d1", Name: "onAdd", File: "src/checkout.ts", Line: 3},
		{Kind: "d2", Name: "Page › render", File: "src/checkout.ts", Line: 12},
		{Kind: "test", Name: "cart › formats", File: "src/cart.spec.ts", Line: 5},
	}, report.confirmed(), "nearest dependents first, then the tests, without the file in the name")

	assert.Equal(t, []consumerRow{
		{Kind: "?", Name: "register", File: "src/events.ts", Line: 9, Detail: "value_ref at line 9"},
		{Kind: "?", Name: "wire", File: "src/events.ts", Line: 20, Detail: "value_ref at line 21\nvalue_ref at line 25"},
	}, report.possible(), "a single site is where the caller is opened")
}

func TestParseConsumersRejectsOtherOutput(t *testing.T) {
	_, err := parseConsumers("not json")
	assert.Error(t, err)
}

func TestConsumerRowLocation(t *testing.T) {
	assert.Equal(t, "src/a.ts:7", consumerRow{File: "src/a.ts", Line: 7}.location())
}

func TestBreadcrumbWithoutFile(t *testing.T) {
	assert.Equal(t, "Cart › addItem", breadcrumbWithoutFile([]string{"src/cart.ts", "Cart", "addItem"}, "addItem"))
	assert.Equal(t, "imports", breadcrumbWithoutFile([]string{"imports"}, "x"), "a lone element is not a file")
	assert.Equal(t, "fallback", breadcrumbWithoutFile(nil, "fallback"))
}

func TestNextConsumersDepth(t *testing.T) {
	assert.Equal(t, 2, nextConsumersDepth(1))
	assert.Equal(t, 0, nextConsumersDepth(2))
	assert.Equal(t, 1, nextConsumersDepth(0))
}
