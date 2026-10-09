package service

import (
	"AB_system/internal/domain/decision"
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/targeting"
	"AB_system/pkg/errs"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Источник значения в ответе Decide.
const (
	SourceDefault     = "default"      // эксперимента нет или пользователь не подошёл
	SourceExperiment  = "experiment"   // значение из варианта эксперимента
	SourceUnknownFlag = "unknown_flag" // флага с таким ключом нет
)

// Узкие интерфейсы: сервису нужны только эти методы репозиториев.
type flagByKeyFinder interface {
	GetFeatureFlagByKey(ctx context.Context, key string) (models.FeatureFlag, error)
}

type runningExperimentFinder interface {
	GetRunningExperimentByFlagID(ctx context.Context, flagID uuid.UUID) (models.Experiment, error)
}

// decisionRecorder сохраняет решения, чтобы события могли ссылаться на decision_id.
type decisionRecorder interface {
	RecordDecisions(ctx context.Context, decisions []models.Decision) error
}

type DecideService struct {
	flags       flagByKeyFinder
	experiments runningExperimentFinder
	recorder    decisionRecorder
	now         func() time.Time
}

func NewDecideService(flags flagByKeyFinder, experiments runningExperimentFinder, recorder decisionRecorder) *DecideService {
	return &DecideService{flags: flags, experiments: experiments, recorder: recorder, now: time.Now}
}

type DecideInput struct {
	SubjectID  string
	Attributes map[string]string
	FlagKeys   []string
}

type DecideItem struct {
	FlagKey      string
	Value        *string // nil для unknown_flag
	ValueType    string
	Source       string
	DecisionID   *uuid.UUID // только для source == experiment
	ExperimentID *uuid.UUID
	Variant      string
}

// Decide возвращает решение по каждому запрошенному флагу, в порядке запроса.
func (s *DecideService) Decide(ctx context.Context, in DecideInput) ([]DecideItem, error) {
	items := make([]DecideItem, 0, len(in.FlagKeys))
	var decisions []models.Decision
	for _, key := range in.FlagKeys {
		item, rec, err := s.decideOne(ctx, in, key)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if rec != nil {
			decisions = append(decisions, *rec)
		}
	}
	// решения записываем до ответа: клиент пришлёт события с этим decision_id
	if err := s.recorder.RecordDecisions(ctx, decisions); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *DecideService) decideOne(ctx context.Context, in DecideInput, key string) (DecideItem, *models.Decision, error) {
	flag, err := s.flags.GetFeatureFlagByKey(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return DecideItem{FlagKey: key, Source: SourceUnknownFlag}, nil, nil
		}
		return DecideItem{}, nil, err
	}

	def := DecideItem{FlagKey: key, Value: &flag.DefaultValue, ValueType: flag.ValueType, Source: SourceDefault}

	exp, err := s.experiments.GetRunningExperimentByFlagID(ctx, flag.ID)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return def, nil, nil // активного эксперимента нет
		}
		return DecideItem{}, nil, err
	}

	if !matchesTargeting(ctx, exp, in.Attributes) {
		return def, nil, nil
	}

	bucket := decision.Bucket(exp.ID, in.SubjectID)
	variant, ok := decision.Select(exp.Variants, exp.AudienceBP, bucket)
	if !ok {
		return def, nil, nil // вне аудитории эксперимента
	}

	decisionID := decision.ID(exp.ID, exp.Version, in.SubjectID, key)
	expID := exp.ID
	value := variant.Value

	var attrs json.RawMessage
	if len(in.Attributes) > 0 {
		attrs, _ = json.Marshal(in.Attributes)
	}
	rec := &models.Decision{
		DecisionID: decisionID, ExperimentID: exp.ID, ExperimentVersion: exp.Version,
		FlagKey: key, SubjectID: in.SubjectID, Variant: variant.Name,
		Attributes: attrs, DecidedAt: s.now().UTC(),
	}
	return DecideItem{
		FlagKey:      key,
		Value:        &value,
		ValueType:    flag.ValueType,
		Source:       SourceExperiment,
		DecisionID:   &decisionID,
		ExperimentID: &expID,
		Variant:      variant.Name,
	}, rec, nil
}

// matchesTargeting проверяет, подходит ли пользователь под правило эксперимента.
// Пустое правило подходит всем. Битое правило (не должно случаться: оно проверяется
// при сохранении) закрывает эксперимент для всех и пишется в лог.
func matchesTargeting(ctx context.Context, exp models.Experiment, attrs map[string]string) bool {
	rule, err := targeting.Parse(exp.Targeting)
	if err != nil {
		slog.WarnContext(ctx, "invalid stored targeting rule", "experiment_id", exp.ID, "err", err)
		return false
	}
	return rule.Matches(attrs)
}
