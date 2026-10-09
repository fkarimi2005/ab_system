// Package targeting — DSL правил таргетинга: кому показывать эксперимент.
package targeting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const maxDepth = 10

type Kind int

const (
	KindString Kind = iota
	KindVersion
)

// Attributes — разрешённые атрибуты. Новый атрибут добавляется здесь, код правил не меняется.
var Attributes = map[string]Kind{
	"country":  KindString,
	"platform": KindString,
	"version":  KindVersion,
}

// Node — узел дерева: ровно одно из And / Or / Not либо лист (Attr + Op + Value).
type Node struct {
	And []Node `json:"and,omitempty"`
	Or  []Node `json:"or,omitempty"`
	Not *Node  `json:"not,omitempty"`

	Attr  string `json:"attr,omitempty"`
	Op    string `json:"op,omitempty"`
	Value any    `json:"value,omitempty"`
}

func Parse(raw []byte) (*Node, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var n Node
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&n); err != nil {
		return nil, fmt.Errorf("targeting: %w", err)
	}
	if err := n.validate(1); err != nil {
		return nil, fmt.Errorf("targeting: %w", err)
	}
	return &n, nil
}

func (n *Node) validate(depth int) error {
	if depth > maxDepth {
		return errors.New("слишком глубокая вложенность")
	}

	kinds := 0
	if n.And != nil {
		kinds++
	}
	if n.Or != nil {
		kinds++
	}
	if n.Not != nil {
		kinds++
	}
	if n.Attr != "" || n.Op != "" || n.Value != nil {
		kinds++
	}
	if kinds != 1 {
		return errors.New("узел должен содержать ровно одно из: and, or, not или условие attr/op/value")
	}

	switch {
	case n.And != nil:
		return validateChildren("and", n.And, depth)
	case n.Or != nil:
		return validateChildren("or", n.Or, depth)
	case n.Not != nil:
		return n.Not.validate(depth + 1)
	default:
		return n.validateLeaf()
	}
}

func validateChildren(name string, children []Node, depth int) error {
	if len(children) == 0 {
		return fmt.Errorf("%s: список условий пуст", name)
	}
	for i := range children {
		if err := children[i].validate(depth + 1); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) validateLeaf() error {
	kind, ok := Attributes[n.Attr]
	if !ok {
		return fmt.Errorf("неизвестный атрибут %q", n.Attr)
	}
	switch n.Op {
	case "==", "!=":
		return checkValue(n.Attr, n.Value, kind)
	case "in":
		list, ok := n.Value.([]any)
		if !ok || len(list) == 0 {
			return fmt.Errorf("%s: для in нужен непустой массив", n.Attr)
		}
		for _, v := range list {
			if err := checkValue(n.Attr, v, kind); err != nil {
				return err
			}
		}
		return nil
	case ">", ">=", "<", "<=":
		if kind != KindVersion {
			return fmt.Errorf("%s: оператор %s не поддерживается для этого атрибута", n.Attr, n.Op)
		}
		return checkValue(n.Attr, n.Value, kind)
	default:
		return fmt.Errorf("неизвестный оператор %q", n.Op)
	}
}

func checkValue(attr string, v any, kind Kind) error {
	s, ok := v.(string)
	if !ok || s == "" {
		return fmt.Errorf("%s: значение должно быть непустой строкой", attr)
	}
	if kind == KindVersion {
		if _, err := parseVersion(s); err != nil {
			return fmt.Errorf("%s: %w", attr, err)
		}
	}
	return nil
}
