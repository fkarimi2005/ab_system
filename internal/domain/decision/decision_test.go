package decision

import (
	"math"
	"strconv"
	"testing"

	"AB_system/internal/domain/models"

	"github.com/google/uuid"
)

var expID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func variants() []models.ExperimentVariant {
	// аудитория 20% = 2000 из 10000, вес поровну
	return []models.ExperimentVariant{
		{Name: "red", Value: "red", Weight: 1000},
		{Name: "blue", Value: "blue", Weight: 1000, IsControl: true},
	}
}

func TestBucketDeterministicAndInRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := "user-" + strconv.Itoa(i)
		a, b := Bucket(expID, s), Bucket(expID, s)
		if a != b {
			t.Fatalf("%s: нестабильная корзина %d != %d", s, a, b)
		}
		if a < 0 || a >= models.FullAudienceBP {
			t.Fatalf("%s: корзина %d вне диапазона", s, a)
		}
	}
}

func TestBucketDependsOnExperiment(t *testing.T) {
	other := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	diff := 0
	for i := 0; i < 1000; i++ {
		s := "user-" + strconv.Itoa(i)
		if Bucket(expID, s) != Bucket(other, s) {
			diff++
		}
	}
	if diff < 900 {
		t.Errorf("корзины в разных экспериментах почти совпадают: различий %d из 1000", diff)
	}
}

func TestSelectBoundaries(t *testing.T) {
	v := variants() // по имени: blue [0,1000), red [1000,2000)
	cases := []struct {
		bucket int
		want   string
		ok     bool
	}{
		{0, "blue", true}, {999, "blue", true},
		{1000, "red", true}, {1999, "red", true},
		{2000, "", false}, {9999, "", false},
	}
	for _, c := range cases {
		got, ok := Select(v, 2000, c.bucket)
		if ok != c.ok || got.Name != c.want {
			t.Errorf("bucket %d: got (%q,%v), want (%q,%v)", c.bucket, got.Name, ok, c.want, c.ok)
		}
	}
}

func TestSelectIgnoresInputOrder(t *testing.T) {
	a := variants()
	b := []models.ExperimentVariant{a[1], a[0]}
	for bucket := 0; bucket < 2000; bucket += 37 {
		x, _ := Select(a, 2000, bucket)
		y, _ := Select(b, 2000, bucket)
		if x.Name != y.Name {
			t.Fatalf("bucket %d: порядок вариантов влияет на результат", bucket)
		}
	}
}

func TestSelectWeightsDoNotCoverAudience(t *testing.T) {
	if _, ok := Select([]models.ExperimentVariant{{Name: "a", Weight: 500}}, 2000, 1500); ok {
		t.Error("при несогласованных весах вариант не должен выбираться")
	}
}

func TestDistributionMatchesWeights(t *testing.T) {
	const n = 100000
	counts := map[string]int{}
	out := 0
	for i := 0; i < n; i++ {
		b := Bucket(expID, "user-"+strconv.Itoa(i))
		if v, ok := Select(variants(), 2000, b); ok {
			counts[v.Name]++
		} else {
			out++
		}
	}
	check := func(name string, got int, want float64) {
		if math.Abs(float64(got)/n-want) > 0.01 {
			t.Errorf("%s: доля %.3f, ожидалось %.3f ±0.01", name, float64(got)/n, want)
		}
	}
	check("blue", counts["blue"], 0.10)
	check("red", counts["red"], 0.10)
	check("вне аудитории", out, 0.80)
}

func TestIDDeterministic(t *testing.T) {
	a := ID(expID, 2, "user-1", "button")
	if a != ID(expID, 2, "user-1", "button") {
		t.Error("ID нестабилен")
	}
	if a == ID(expID, 3, "user-1", "button") || a == ID(expID, 2, "user-2", "button") || a == ID(expID, 2, "user-1", "other") {
		t.Error("ID не зависит от одного из входов")
	}
}
