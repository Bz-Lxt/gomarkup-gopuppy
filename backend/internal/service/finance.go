package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"gopuppy/internal/clock"
	"gopuppy/internal/domain"
	"gopuppy/internal/repo"
)

type Finance struct {
	Repo   *repo.Finance
	Family *Family
}

type WeightInput struct {
	WeightKG   float64 `json:"weight_kg"`
	MeasuredAt string  `json:"measured_at"`
	Note       string  `json:"note"`
}

type ExpenseInput struct {
	Category    domain.ExpenseCategory `json:"category"`
	AmountCents int64                  `json:"amount_cents"`
	SpentAt     string                 `json:"spent_at"`
	Note        string                 `json:"note"`
}

func (s *Finance) AddWeight(ctx context.Context, userID, petID uuid.UUID, in WeightInput) (*domain.WeightRecord, error) {
	if _, _, err := s.Family.MustWritePet(ctx, userID, petID); err != nil {
		return nil, err
	}
	if in.WeightKG <= 0 || in.WeightKG > 200 {
		return nil, fmt.Errorf("%w: weight_kg", domain.ErrValidation)
	}
	at, err := parseDateTime(in.MeasuredAt)
	if err != nil {
		at = clock.Now()
	}
	w := &domain.WeightRecord{ID: uuid.New(), PetID: petID, WeightKG: in.WeightKG, MeasuredAt: at, Note: in.Note, CreatedBy: userID}
	if err := s.Repo.AddWeight(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Finance) AddExpense(ctx context.Context, userID, petID uuid.UUID, in ExpenseInput) (*domain.Expense, error) {
	if _, _, err := s.Family.MustWritePet(ctx, userID, petID); err != nil {
		return nil, err
	}
	if !in.Category.Valid() || in.AmountCents < 0 {
		return nil, fmt.Errorf("%w: category/amount", domain.ErrValidation)
	}
	at, err := parseDateTime(in.SpentAt)
	if err != nil {
		at = clock.Now()
	}
	e := &domain.Expense{ID: uuid.New(), PetID: petID, Category: in.Category, AmountCents: in.AmountCents, SpentAt: at, Note: in.Note, CreatedBy: userID}
	if err := s.Repo.AddExpense(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Finance) Summary(ctx context.Context, userID, petID uuid.UUID) (*domain.FinanceSummary, error) {
	p, _, err := s.Family.MustReadPet(ctx, userID, petID)
	if err != nil {
		return nil, err
	}
	since := repo.MonthStart12()
	weights, err := s.Repo.WeightsSince(ctx, petID, since)
	if err != nil {
		return nil, err
	}
	exps, err := s.Repo.ExpensesSince(ctx, petID, since)
	if err != nil {
		return nil, err
	}
	sum := &domain.FinanceSummary{WeightMin: p.WeightMin, WeightMax: p.WeightMax}
	monthKey := clock.Today().Format("2006-01")
	year := clock.Today().Year()
	catYear := map[string]int64{}
	monthBuckets := map[string]*domain.ExpenseMonthBucket{}
	for i := 0; i < 12; i++ {
		m := since.AddDate(0, i, 0).Format("2006-01")
		monthBuckets[m] = &domain.ExpenseMonthBucket{Month: m, ByCat: map[string]int64{}}
	}
	for _, e := range exps {
		mk := e.SpentAt.In(clock.Beijing).Format("2006-01")
		b := monthBuckets[mk]
		if b == nil {
			continue
		}
		b.ByCat[string(e.Category)] += e.AmountCents
		b.Total += e.AmountCents
		if e.SpentAt.In(clock.Beijing).Year() == year {
			sum.YearTotalCents += e.AmountCents
			catYear[string(e.Category)] += e.AmountCents
		}
		if mk == monthKey {
			sum.MonthTotalCents += e.AmountCents
		}
	}
	for i := 0; i < 12; i++ {
		m := since.AddDate(0, i, 0).Format("2006-01")
		sum.ExpenseSeries = append(sum.ExpenseSeries, *monthBuckets[m])
	}
	wMonth := map[string][]float64{}
	for _, w := range weights {
		mk := w.MeasuredAt.In(clock.Beijing).Format("2006-01")
		wMonth[mk] = append(wMonth[mk], w.WeightKG)
	}
	var prev *float64
	for i := 0; i < 12; i++ {
		m := since.AddDate(0, i, 0).Format("2006-01")
		vals := wMonth[m]
		pt := domain.WeightPoint{Month: m}
		if len(vals) > 0 {
			min, max, acc := vals[0], vals[0], 0.0
			for _, v := range vals {
				acc += v
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
			pt.AvgKG = acc / float64(len(vals))
			pt.MinKG = min
			pt.MaxKG = max
			if prev != nil && *prev > 0 && abs((pt.AvgKG-*prev)/ *prev) > 0.10 {
				pt.Anomaly = true
			}
			v := pt.AvgKG
			prev = &v
		}
		sum.WeightSeries = append(sum.WeightSeries, pt)
	}
	var shares []domain.CategoryShare
	var total int64
	for k, v := range catYear {
		shares = append(shares, domain.CategoryShare{Category: k, Cents: v})
		total += v
	}
	sort.Slice(shares, func(i, j int) bool { return shares[i].Cents > shares[j].Cents })
	for i := range shares {
		if total > 0 {
			shares[i].Percent = float64(shares[i].Cents) * 100 / float64(total)
		}
	}
	sum.Pie = shares
	if len(shares) > 3 {
		sum.Top3 = shares[:3]
	} else {
		sum.Top3 = shares
	}
	return sum, nil
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
