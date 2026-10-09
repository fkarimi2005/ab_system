package targeting

import (
	"fmt"
	"strconv"
	"strings"
)

// parseVersion превращает "2.10.1" в []int{2, 10, 1}.
func parseVersion(s string) ([]int, error) {
	parts := strings.Split(s, ".")
	res := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("некорректная версия %q", s)
		}
		res = append(res, n)
	}
	return res, nil
}

// compareVersions сравнивает по сегментам; недостающие сегменты считаются нулями
// (поэтому 2.5 == 2.5.0). Возвращает -1, 0 или 1.
func compareVersions(a, b []int) int {
	for i := 0; i < max(len(a), len(b)); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}
