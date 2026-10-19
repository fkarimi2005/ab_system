package targeting

func (n *Node) Matches(attrs map[string]string) bool {
	if n == nil {
		return true
	}
	switch {
	case n.And != nil:
		for i := range n.And {
			if !n.And[i].Matches(attrs) {
				return false
			}
		}
		return true
	case n.Or != nil:
		for i := range n.Or {
			if n.Or[i].Matches(attrs) {
				return true
			}
		}
		return false
	case n.Not != nil:
		return !n.Not.Matches(attrs)
	default:
		return n.matchLeaf(attrs)
	}
}

func (n *Node) matchLeaf(attrs map[string]string) bool {
	actual, ok := attrs[n.Attr]
	if !ok {
		return false
	}
	kind := Attributes[n.Attr]

	switch n.Op {
	case "in":
		for _, v := range n.Value.([]any) {
			if equal(kind, actual, v.(string)) {
				return true
			}
		}
		return false
	case "==":
		return equal(kind, actual, n.Value.(string))
	case "!=":
		return !equal(kind, actual, n.Value.(string))
	}

	a, err := parseVersion(actual)
	if err != nil {
		return false
	}
	b, _ := parseVersion(n.Value.(string))
	c := compareVersions(a, b)
	switch n.Op {
	case ">":
		return c > 0
	case ">=":
		return c >= 0
	case "<":
		return c < 0
	default:
		return c <= 0
	}
}

func equal(kind Kind, actual, want string) bool {
	if kind == KindVersion {
		a, err := parseVersion(actual)
		if err != nil {
			return false
		}
		b, _ := parseVersion(want)
		return compareVersions(a, b) == 0
	}
	return actual == want
}
