package service

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/models/seeds"
	"AB_system/internal/domain/repository"
	"context"
	"github.com/google/uuid"
)

type ReportService struct {
	reportRepo     repository.ReportRepository
	experimentRepo repository.ExperimentRepository
	event          repository.EventRepository

	metricRepo repository.MetricRepository
}

func NewReportService(reportRepo repository.ReportRepository,
	exper repository.ExperimentRepository,
	event repository.EventRepository,
	m repository.MetricRepository,
) *ReportService {
	return &ReportService{reportRepo: reportRepo,
		experimentRepo: exper,
		event:          event,

		metricRepo: m,
	}
}

type ExperimentWithVariant struct {
	ExperimentID uuid.UUID
	VariantID    []uuid.UUID
}

func (s *ReportService) Update(ctx context.Context) error {
	experiments, err := s.experimentRepo.GetAllExperiments(ctx)
	if err != nil {
		return err
	}
	expMap := make(map[uuid.UUID]*ExperimentWithVariant, len(experiments))
	for _, e := range experiments {

		exp, ok := expMap[e.ID]
		if !ok {
			exp = &ExperimentWithVariant{
				ExperimentID: e.ID,
				VariantID:    []uuid.UUID{},
			}
			expMap[e.ID] = exp
		}
		for _, v := range e.Variants {
			exp.VariantID = append(exp.VariantID, v.ID)
		}
	}
	result := make([]ExperimentWithVariant, 0, len(expMap))
	for _, exp := range expMap {
		result = append(result, *exp)
	}
	metrics, err := s.metricRepo.ListMetrics(ctx, true)
	if err != nil {
		return err
	}

	for _, exp := range result {
		for _, variantID := range exp.VariantID {
			showCount, err := s.event.CountEventsByEventType(
				ctx, "shown", exp.ExperimentID, variantID,
			)
			if err != nil {
				return err
			}

			clickCount, err := s.event.CountEventsByEventType(
				ctx, "click", exp.ExperimentID, variantID,
			)
			if err != nil {
				return err
			}

			purchaseCount, err := s.event.CountEventsByEventType(
				ctx, "purchase", exp.ExperimentID, variantID,
			)
			if err != nil {
				return err
			}

			errorCount, err := s.event.CountEventsByEventType(
				ctx, "error", exp.ExperimentID, variantID,
			)

			if err != nil {
				return err
			}

			for _, m := range metrics {
				switch seeds.MetricType(m.Key) {

				case seeds.ShownCount:
					err = s.reportRepo.Save(
						ctx, &models.Report{
							ExperimentID: exp.ExperimentID,
							VariantID:    variantID,
							MetricID:     m.ID,
							Count:        float64(showCount)})
					if err != nil {
						return err
					}

				case seeds.ClickCount:
					err = s.reportRepo.Save(ctx,
						&models.Report{ExperimentID: exp.ExperimentID,
							MetricID: m.ID, VariantID: variantID, Count: float64(clickCount)})
					if err != nil {
						return err
					}
				case seeds.ErrorCount:
					err = s.reportRepo.Save(
						ctx, &models.Report{
							ExperimentID: exp.ExperimentID,
							MetricID:     m.ID, VariantID: variantID,
							Count: float64(errorCount),
						})
					if err != nil {
						return err
					}

				}

			}
			APICount, err := s.event.CountEventsByEventType(ctx, "api_request", exp.ExperimentID, variantID)
			if err != nil {
				return err
			}

			ctr := 0.0
			if showCount > 0 {
				ctr = float64(clickCount) / float64(showCount)
			}

			purchaseRate := 0.0
			if clickCount > 0 {
				purchaseRate = float64(purchaseCount) / float64(clickCount)
			}
			errorRate := 0.0
			if APICount > 0 {
				errorRate = float64(errorCount) / float64(APICount)
			}
			for _, m := range metrics {
				switch seeds.MetricType(m.Key) {
				case seeds.CTR:
					err = s.reportRepo.Save(ctx, &models.Report{
						ExperimentID: exp.ExperimentID,
						VariantID:    variantID,
						MetricID:     m.ID,
						Count:        ctr,
					})
					if err != nil {
						return err
					}

				case seeds.PurchaseRate:
					err = s.reportRepo.Save(ctx, &models.Report{
						ExperimentID: exp.ExperimentID,
						VariantID:    variantID,
						MetricID:     m.ID,
						Count:        purchaseRate,
					})
					if err != nil {
						return err
					}
				case seeds.APIRequestCount:
					err = s.reportRepo.Save(
						ctx, &models.Report{
							ExperimentID: exp.ExperimentID,
							MetricID:     m.ID, VariantID: variantID,
							Count: float64(APICount),
						})
					if err != nil {
						return err
					}
				case seeds.ErrorRate:
					err = s.reportRepo.Save(ctx, &models.Report{
						ExperimentID: exp.ExperimentID,
						MetricID:     m.ID, VariantID: variantID,
						Count: errorRate,
					})
					if err != nil {
						return err
					}
				}

			}
		}
	}

	return nil
}
func (s *ReportService) Reports(
	ctx context.Context,
) ([]models.ReportResult, error) {
	report, err := s.reportRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	reportMap := make(map[uuid.UUID]*models.ReportResult)

	variantMaps := make(map[uuid.UUID]map[uuid.UUID]*models.VariantReport)

	metricCache := make(map[uuid.UUID]string)

	for _, r := range report {
		re, ok := reportMap[r.ExperimentID]
		if !ok {
			re = &models.ReportResult{
				ExperimentID:  r.ExperimentID,
				VariantReport: []models.VariantReport{},
			}

			reportMap[r.ExperimentID] = re
			variantMaps[r.ExperimentID] =
				make(map[uuid.UUID]*models.VariantReport)
		}

		variantMap := variantMaps[r.ExperimentID]

		va, ok := variantMap[r.VariantID]
		if !ok {
			va = &models.VariantReport{
				VariantID: r.VariantID,
				Metrics:   []models.MetricWithCount{},
			}

			variantMap[r.VariantID] = va
		}

		metricKey, ok := metricCache[r.MetricID]
		if !ok {
			metric, err := s.metricRepo.GetMetricByID(ctx, r.MetricID)
			if err != nil {
				return nil, err
			}

			metricKey = metric.Key
			metricCache[r.MetricID] = metricKey
		}

		va.Metrics = append(va.Metrics, models.MetricWithCount{
			MetricKey: metricKey,
			Count:     r.Count,
		})
	}

	reports := make([]models.ReportResult, 0, len(reportMap))

	for experimentID, re := range reportMap {
		for _, va := range variantMaps[experimentID] {
			re.VariantReport = append(
				re.VariantReport,
				*va,
			)
		}

		reports = append(reports, *re)
	}

	return reports, nil
}
func (s *ReportService) Report(
	ctx context.Context,
	experimentID uuid.UUID,
) (models.ReportResult, error) {
	report, err := s.reportRepo.Get(ctx, experimentID)
	if err != nil {
		return models.ReportResult{}, err
	}

	reportMap := make(map[uuid.UUID]*models.ReportResult)

	variantMaps := make(map[uuid.UUID]map[uuid.UUID]*models.VariantReport)

	metricCache := make(map[uuid.UUID]string)

	for _, r := range report {
		re, ok := reportMap[r.ExperimentID]
		if !ok {
			re = &models.ReportResult{
				ExperimentID:  r.ExperimentID,
				VariantReport: []models.VariantReport{},
			}

			reportMap[r.ExperimentID] = re
			variantMaps[r.ExperimentID] =
				make(map[uuid.UUID]*models.VariantReport)
		}

		variantMap := variantMaps[r.ExperimentID]

		va, ok := variantMap[r.VariantID]
		if !ok {
			va = &models.VariantReport{
				VariantID: r.VariantID,
				Metrics:   []models.MetricWithCount{},
			}

			variantMap[r.VariantID] = va
		}

		metricKey, ok := metricCache[r.MetricID]
		if !ok {
			metric, err := s.metricRepo.GetMetricByID(ctx, r.MetricID)
			if err != nil {
				return models.ReportResult{}, err
			}

			metricKey = metric.Key
			metricCache[r.MetricID] = metricKey
		}

		va.Metrics = append(va.Metrics, models.MetricWithCount{
			MetricKey: metricKey,
			Count:     r.Count,
		})
	}

	var reports models.ReportResult

	for experimentID, re := range reportMap {
		for _, va := range variantMaps[experimentID] {
			re.VariantReport = append(
				re.VariantReport,
				*va,
			)
		}

		reports = models.ReportResult{
			ExperimentID:  experimentID,
			VariantReport: re.VariantReport,
		}
	}

	return reports, nil
}
