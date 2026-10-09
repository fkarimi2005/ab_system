package service

import (
	"AB_system/internal/domain/models"
	"AB_system/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

type fakeFlags map[string]models.FeatureFlag

func (f fakeFlags) GetFeatureFlagByKey(_ context.Context, key string) (models.FeatureFlag, error) {
	fl, ok := f[key]
	if !ok {
		return models.FeatureFlag{}, errs.ErrRecordNotFound
	}
	return fl, nil
}

type fakeRecorder struct{ saved []models.Decision }

func (f *fakeRecorder) RecordDecisions(_ context.Context, d []models.Decision) error {
	f.saved = append(f.saved, d...)
	return nil
}

type fakeExperiments map[uuid.UUID]models.Experiment // по id флага

func (f fakeExperiments) GetRunningExperimentByFlagID(_ context.Context, id uuid.UUID) (models.Experiment, error) {
	e, ok := f[id]
	if !ok {
		return models.Experiment{}, errs.ErrRecordNotFound
	}
	return e, nil
}

func newDecideFixture(audienceBP int, targetingRule string) (*DecideService, models.Experiment) {
	flag := models.FeatureFlag{ID: uuid.New(), Key: "button", ValueType: "string", DefaultValue: "green"}
	exp := models.Experiment{
		ID: uuid.New(), FeatureFlagID: flag.ID, Status: models.ExperimentStatusRunning,
		Version: 1, AudienceBP: audienceBP,
		Variants: []models.ExperimentVariant{
			{Name: "blue", Value: "blue", Weight: audienceBP / 2, IsControl: true},
			{Name: "red", Value: "red", Weight: audienceBP / 2},
		},
	}
	if targetingRule != "" {
		exp.Targeting = json.RawMessage(targetingRule)
	}
	idle := models.FeatureFlag{ID: uuid.New(), Key: "idle", ValueType: "bool", DefaultValue: "false"}
	return NewDecideService(
		fakeFlags{"button": flag, "idle": idle},
		fakeExperiments{flag.ID: exp},
		&fakeRecorder{},
	), exp
}

func decideOne(t *testing.T, svc *DecideService, subject string, attrs map[string]string) DecideItem {
	t.Helper()
	items, err := svc.Decide(context.Background(), DecideInput{SubjectID: subject, Attributes: attrs, FlagKeys: []string{"button"}})
	if err != nil {
		t.Fatal(err)
	}
	return items[0]
}

func TestDecideUnknownFlagAndDefault(t *testing.T) {
	svc, _ := newDecideFixture(10000, "")
	items, err := svc.Decide(context.Background(), DecideInput{SubjectID: "u1", FlagKeys: []string{"nope", "idle"}})
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Source != SourceUnknownFlag || items[0].Value != nil {
		t.Errorf("unknown flag: %+v", items[0])
	}
	if items[1].Source != SourceDefault || *items[1].Value != "false" || items[1].DecisionID != nil {
		t.Errorf("flag without experiment: %+v", items[1])
	}
}

func TestDecideDeterministicAndSticky(t *testing.T) {
	svc, exp := newDecideFixture(10000, "")
	first := decideOne(t, svc, "user-42", nil)
	if first.Source != SourceExperiment || first.DecisionID == nil || *first.ExperimentID != exp.ID {
		t.Fatalf("ожидался эксперимент: %+v", first)
	}
	for i := 0; i < 50; i++ {
		again := decideOne(t, svc, "user-42", nil)
		if *again.Value != *first.Value || *again.DecisionID != *first.DecisionID {
			t.Fatal("повторный запрос дал другой результат")
		}
	}
}

func TestDecideAudienceFraction(t *testing.T) {
	svc, _ := newDecideFixture(2000, "")
	in := 0
	for i := 0; i < 20000; i++ {
		it := decideOne(t, svc, "u"+strconv.Itoa(i), nil)
		if it.Source == SourceExperiment {
			in++
		} else if *it.Value != "green" {
			t.Fatal("вне аудитории должно быть значение по умолчанию")
		}
	}
	if share := float64(in) / 20000; share < 0.19 || share > 0.21 {
		t.Errorf("доля аудитории %.3f, ожидалось ~0.20", share)
	}
}

func TestDecideTargeting(t *testing.T) {
	rule := `{"and":[{"attr":"country","op":"in","value":["RU","KZ"]},{"attr":"version","op":">=","value":"2.5.0"}]}`
	svc, _ := newDecideFixture(10000, rule) // аудитория 100%, поэтому решает только таргетинг

	if it := decideOne(t, svc, "u1", map[string]string{"country": "RU", "version": "2.10.0"}); it.Source != SourceExperiment {
		t.Errorf("подходящий пользователь не в эксперименте: %+v", it)
	}
	for name, attrs := range map[string]map[string]string{
		"другая страна": {"country": "US", "version": "3.0.0"},
		"старая версия": {"country": "RU", "version": "2.4.9"},
		"нет атрибутов": nil,
	} {
		it := decideOne(t, svc, "u1", attrs)
		if it.Source != SourceDefault || *it.Value != "green" || it.DecisionID != nil {
			t.Errorf("%s: должен получить default, got %+v", name, it)
		}
	}
}

func TestDecideBrokenStoredRuleFailsClosed(t *testing.T) {
	svc, _ := newDecideFixture(10000, `{"attr":"age","op":"==","value":"1"}`)
	if it := decideOne(t, svc, "u1", map[string]string{"age": "1"}); it.Source != SourceDefault {
		t.Errorf("битое правило должно закрывать эксперимент: %+v", it)
	}
}

func TestNormalizeTargeting(t *testing.T) {
	for _, empty := range []string{"", "null", "  "} {
		got, err := normalizeTargeting([]byte(empty))
		if err != nil || got != nil {
			t.Errorf("%q: got %q, %v; ожидалось nil без ошибки", empty, got, err)
		}
	}
	if got, err := normalizeTargeting([]byte(`{"attr":"country","op":"==","value":"RU"}`)); err != nil || len(got) == 0 {
		t.Errorf("корректное правило: %q, %v", got, err)
	}
	if _, err := normalizeTargeting([]byte(`{"attr":"age","op":"==","value":"1"}`)); err == nil || !isInvalidTargeting(err) {
		t.Errorf("ожидалась ErrInvalidTargeting, got %v", err)
	}
}

func isInvalidTargeting(err error) bool { return errors.Is(err, errs.ErrInvalidTargeting) }

func TestDecideRecordsOnlyExperimentDecisions(t *testing.T) {
	svc, exp := newDecideFixture(10000, "")
	rec := svc.recorder.(*fakeRecorder)

	// флаг без эксперимента, неизвестный флаг и вне аудитории не записываются
	if _, err := svc.Decide(context.Background(), DecideInput{SubjectID: "u1", FlagKeys: []string{"idle", "nope"}}); err != nil {
		t.Fatal(err)
	}
	if len(rec.saved) != 0 {
		t.Fatalf("записано %d решений, ожидалось 0", len(rec.saved))
	}

	items, err := svc.Decide(context.Background(), DecideInput{
		SubjectID: "u1", Attributes: map[string]string{"country": "RU"}, FlagKeys: []string{"button"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.saved) != 1 {
		t.Fatalf("записано %d решений, ожидалось 1", len(rec.saved))
	}
	d := rec.saved[0]
	if d.DecisionID != *items[0].DecisionID || d.ExperimentID != exp.ID || d.SubjectID != "u1" ||
		d.Variant != items[0].Variant || d.FlagKey != "button" || string(d.Attributes) != `{"country":"RU"}` {
		t.Errorf("запись решения не совпадает с ответом: %+v", d)
	}
}
