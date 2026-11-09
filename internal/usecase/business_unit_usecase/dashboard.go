package business_unit_usecase

import (
	"context"
	"database/sql"

	"ppl-m1-be/internal/usecase/apperr"
	"ppl-m1-be/internal/usecase/repository/sqlc"
)

// Metric adalah satu data yang dialirkan modul ke dashboard. Value ditambahkan
// ke akumulasi harian; PeriodDate kosong berarti hari ini (WIB). BusinessUnitID
// kosong untuk penjualan partner yang tidak berada di bawah unit usaha.
type Metric struct {
	OrganizationID int64
	BusinessUnitID sql.NullInt64
	SourceModule   string
	Key            string
	Label          string
	Unit           string
	Value          int64
	PeriodDate     sql.NullTime
}

type ReportFilter struct {
	BusinessUnitID sql.NullInt64
	From           sql.NullTime
	To             sql.NullTime
}

// Push dipanggil modul lain dalam transaksi yang sama dengan perubahan
// datanya, sehingga tidak perlu membangun dashboard sendiri.
func Push(ctx context.Context, q *sqlc.Queries, m Metric) error {
	return q.UpsertDashboardMetric(ctx, sqlc.UpsertDashboardMetricParams{
		OrganizationID: m.OrganizationID, BusinessUnitID: m.BusinessUnitID, SourceModule: m.SourceModule,
		MetricKey: m.Key, Label: m.Label, Unit: m.Unit, Value: m.Value, PeriodDate: m.PeriodDate,
	})
}

// Submit menerima data dashboard lewat HTTP.
func (u *Usecase) Submit(ctx context.Context, m Metric) error {
	if _, err := u.Get(ctx, m.OrganizationID, m.BusinessUnitID.Int64); err != nil {
		if err == apperr.BusinessUnitNotFound {
			return apperr.FieldErrors{"business_unit_id": "Unit usaha tidak terdaftar"}
		}
		return err
	}
	return Push(ctx, u.store.Queries, m)
}

func (u *Usecase) Report(ctx context.Context, orgID int64, f ReportFilter) ([]sqlc.ReportDashboardMetricsRow, error) {
	return u.store.ReportDashboardMetrics(ctx, sqlc.ReportDashboardMetricsParams{
		OrganizationID: orgID, BusinessUnitID: f.BusinessUnitID, DateFrom: f.From, DateTo: f.To,
	})
}
