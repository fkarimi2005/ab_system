package targeting

import "testing"

func TestMatches(t *testing.T) {
	rule := `{"and":[
		{"attr":"country","op":"in","value":["RU","KZ"]},
		{"attr":"version","op":">=","value":"2.5.0"},
		{"not":{"attr":"platform","op":"==","value":"web"}}
	]}`
	n, err := Parse([]byte(rule))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		attrs map[string]string
		want  bool
	}{
		{"подходит", map[string]string{"country": "RU", "version": "2.6.0", "platform": "ios"}, true},
		{"другая страна", map[string]string{"country": "US", "version": "2.6.0", "platform": "ios"}, false},
		{"версия ниже", map[string]string{"country": "RU", "version": "2.4.9", "platform": "ios"}, false},
		{"граница версии", map[string]string{"country": "KZ", "version": "2.5.0", "platform": "android"}, true},
		{"2.10 новее 2.9", map[string]string{"country": "RU", "version": "2.10", "platform": "ios"}, true},
		{"web запрещён", map[string]string{"country": "RU", "version": "3.0.0", "platform": "web"}, false},
		{"нет platform: not(...) истинно", map[string]string{"country": "RU", "version": "3.0.0"}, true},
		{"нет country", map[string]string{"version": "3.0.0", "platform": "ios"}, false},
		{"мусор в версии", map[string]string{"country": "RU", "version": "abc", "platform": "ios"}, false},
	}
	for _, tt := range tests {
		if got := n.Matches(tt.attrs); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestEmptyRuleMatchesEveryone(t *testing.T) {
	for _, raw := range []string{"", "  ", "null"} {
		n, err := Parse([]byte(raw))
		if err != nil || n != nil {
			t.Fatalf("%q: n=%v err=%v", raw, n, err)
		}
		if !n.Matches(nil) {
			t.Errorf("%q: nil-правило должно подходить всем", raw)
		}
	}
}

func TestParseErrors(t *testing.T) {
	bad := map[string]string{
		"кривой json":             `{"and":`,
		"неизвестное поле":        `{"xor":[]}`,
		"пустой объект":           `{}`,
		"пустой and":              `{"and":[]}`,
		"неизвестный атрибут":     `{"attr":"age","op":"==","value":"1"}`,
		"неизвестный оператор":    `{"attr":"country","op":"like","value":"R"}`,
		"порядок для строки":      `{"attr":"country","op":">","value":"RU"}`,
		"in без массива":          `{"attr":"country","op":"in","value":"RU"}`,
		"in пустой":               `{"attr":"country","op":"in","value":[]}`,
		"число вместо строки":     `{"attr":"country","op":"==","value":5}`,
		"плохая версия в правиле": `{"attr":"version","op":">","value":"2.x"}`,
		"нет значения":            `{"attr":"country","op":"=="}`,
	}
	for name, raw := range bad {
		if _, err := Parse([]byte(raw)); err == nil {
			t.Errorf("%s: ожидалась ошибка", name)
		}
	}
}

func TestDepthLimit(t *testing.T) {
	raw := `{"attr":"country","op":"==","value":"RU"}`
	for i := 0; i < maxDepth+1; i++ {
		raw = `{"not":` + raw + `}`
	}
	if _, err := Parse([]byte(raw)); err == nil {
		t.Error("ожидалась ошибка глубины")
	}
}
