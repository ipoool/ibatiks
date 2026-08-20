package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/money"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

type ReportService struct {
	pool    *pgxpool.Pool
	reports *repository.ReportRepo
	trips   *repository.TripRepo
	orders  *repository.OrderRepo
}

func NewReportService(
	pool *pgxpool.Pool,
	reports *repository.ReportRepo,
	trips *repository.TripRepo,
	orders *repository.OrderRepo,
) *ReportService {
	return &ReportService{pool: pool, reports: reports, trips: trips, orders: orders}
}

// TripProfit menyusun laporan keuangan satu trip.
//
// Angka mentahnya diambil dari database, tapi penjumlahan akhir dikerjakan di
// sini supaya definisi laba tetap terbaca sebagai kode, bukan tersembunyi di
// dalam SQL yang panjang.
// tripID nil berarti seluruh trip dijumlahkan jadi satu laporan.
func (s *ReportService) TripProfit(ctx context.Context, tripID *uuid.UUID, from, to *time.Time) (*domain.TripProfitReport, error) {
	// Identitas trip hanya diambil kalau memang satu trip yang diminta.
	// Memaksakan pembacaan trip saat laporannya lintas trip berarti satu kueri
	// yang hasilnya dibuang.
	var trip *domain.Trip
	if tripID != nil {
		var err error
		trip, err = s.trips.GetByID(ctx, s.pool, *tripID)
		if err != nil {
			return nil, err
		}
	}

	fin, err := s.reports.TripFinancials(ctx, s.pool, tripID, from, to)
	if err != nil {
		return nil, err
	}
	breakdown, err := s.reports.ExpenseBreakdown(ctx, s.pool, tripID, from, to)
	if err != nil {
		return nil, err
	}
	if breakdown == nil {
		breakdown = []domain.ExpenseBreakdown{}
	}

	// Pengelompokan per trip hanya berarti kalau laporannya memang mencakup
	// lebih dari satu trip.
	byTrip := []domain.TripExpenseBreakdown{}
	if tripID == nil {
		byTrip, err = s.reports.ExpenseBreakdownByTrip(ctx, s.pool, from, to)
		if err != nil {
			return nil, err
		}
	}

	grossProfit := fin.Revenue.Sub(fin.COGS)
	netProfit := grossProfit.Sub(fin.TripExpenses)

	// Margin dihitung terhadap omzet; trip tanpa omzet ditulis 0 alih-alih
	// menghasilkan pembagian dengan nol.
	margin := decimal.Zero
	if fin.Revenue.GreaterThan(decimal.Zero) {
		margin = netProfit.Div(fin.Revenue).Mul(decimal.NewFromInt(100)).Round(2)
	}

	laporan := &domain.TripProfitReport{

		Revenue:      money.RoundRupiah(fin.Revenue),
		COGS:         money.RoundRupiah(fin.COGS),
		GrossProfit:  money.RoundRupiah(grossProfit),
		TripExpenses: money.RoundRupiah(fin.TripExpenses),
		NetProfit:    money.RoundRupiah(netProfit),
		MarginPct:    margin,

		ShippingFeeCollected: money.RoundRupiah(fin.ShippingFeeCollected),
		ShippingCostPaid:     money.RoundRupiah(fin.ShippingCostPaid),
		DiscountGiven:        money.RoundRupiah(fin.DiscountGiven),

		SurplusStockQty:   fin.SurplusStockQty,
		SurplusStockValue: money.RoundRupiah(fin.SurplusStockValue),

		// Uang yang benar-benar keluar selama trip: seluruh belanja (termasuk
		// yang jadi stok) ditambah biaya perjalanan.
		TotalCapitalOut: money.RoundRupiah(fin.PurchaseTotal.Add(fin.TripExpenses)),
		PaymentReceived: money.RoundRupiah(fin.PaymentReceived),
		Outstanding:     money.RoundRupiah(fin.Outstanding),

		OrderCount:    fin.OrderCount,
		CustomerCount: fin.CustomerCount,
		ItemQty:       fin.ItemQty,

		ExpenseBreakdown: breakdown,
		ExpenseByTrip:    byTrip,
	}

	if trip != nil {
		laporan.TripID = &trip.ID
		laporan.TripCode = trip.Code
		laporan.Title = trip.Title
		laporan.Country = trip.Country
		laporan.Status = trip.Status
	}

	return laporan, nil
}

/*
 * TripProfitRows menyusun satu baris laporan untuk tiap trip, dipakai ekspor CSV.
 *
 * Sengaja memanggil TripProfit berulang alih-alih satu kueri yang
 * mengelompokkan per trip: definisi labanya jadi persis sama dengan yang
 * terbaca di layar. Satu kueri gabungan berarti rumus kedua yang harus dijaga
 * tetap sama, dan selisih beberapa rupiah antara layar dan berkas ekspor adalah
 * jenis kesalahan yang paling lama tidak ketahuan.
 *
 * Trip yang tidak punya kegiatan apa pun pada periode itu dilewati — barisnya
 * hanya akan berisi nol dan menutupi trip yang benar-benar berjalan.
 */
func (s *ReportService) TripProfitRows(ctx context.Context, tripID *uuid.UUID, from, to *time.Time) ([]domain.TripProfitReport, error) {
	if tripID != nil {
		laporan, err := s.TripProfit(ctx, tripID, from, to)
		if err != nil {
			return nil, err
		}
		return []domain.TripProfitReport{*laporan}, nil
	}

	trips, _, err := s.trips.List(ctx, s.pool, pagination.Params{Page: 1, PerPage: pagination.ExportPerPage}, "")
	if err != nil {
		return nil, err
	}

	hasil := make([]domain.TripProfitReport, 0, len(trips))
	for i := range trips {
		laporan, err := s.TripProfit(ctx, &trips[i].ID, from, to)
		if err != nil {
			return nil, err
		}
		if laporan.OrderCount == 0 &&
			laporan.TripExpenses.IsZero() &&
			laporan.TotalCapitalOut.IsZero() {
			continue
		}
		hasil = append(hasil, *laporan)
	}
	return hasil, nil
}

func (s *ReportService) OrderProfits(ctx context.Context, p pagination.Params, tripID *uuid.UUID) ([]domain.OrderProfit, int64, error) {
	return s.reports.OrderProfits(ctx, s.pool, p, tripID)
}

func (s *ReportService) Receivables(ctx context.Context, p pagination.Params) ([]domain.Receivable, int64, error) {
	return s.reports.Receivables(ctx, s.pool, p)
}

func (s *ReportService) CustomerSales(ctx context.Context, p pagination.Params, tripID *uuid.UUID) ([]domain.CustomerSales, int64, error) {
	return s.reports.CustomerSales(ctx, s.pool, p, tripID)
}

// CustomerSalesByChannel memecah rekap customer menjadi per kanal, untuk ekspor.
func (s *ReportService) CustomerSalesByChannel(ctx context.Context, tripID *uuid.UUID) ([]domain.CustomerChannelSales, error) {
	return s.reports.CustomerSalesByChannel(ctx, s.pool, tripID)
}

// ChannelSales merangkum penjualan per kanal dan menambahkan porsi omzet tiap
// kanal terhadap total. Porsinya dihitung di sini, bukan di SQL, supaya
// pembulatannya konsisten dengan laporan lain.
func (s *ReportService) ChannelSales(ctx context.Context, tripID *uuid.UUID, from, to *time.Time) ([]domain.ChannelSales, error) {
	rows, err := s.reports.ChannelSales(ctx, s.pool, tripID, from, to)
	if err != nil {
		return nil, err
	}

	totalRevenue := decimal.Zero
	for _, row := range rows {
		totalRevenue = totalRevenue.Add(row.Revenue)
	}

	for i := range rows {
		if totalRevenue.GreaterThan(decimal.Zero) {
			rows[i].RevenueShare = rows[i].Revenue.
				Div(totalRevenue).
				Mul(decimal.NewFromInt(100)).
				Round(1)
		}
	}
	return rows, nil
}

func (s *ReportService) ProductSales(ctx context.Context, limit int, tripID *uuid.UUID, from, to *time.Time) ([]domain.ProductSales, error) {
	if limit < 1 || limit > 200 {
		limit = 20
	}
	return s.reports.ProductSales(ctx, s.pool, limit, tripID, from, to)
}

// Dashboard merangkum kondisi bisnis saat ini untuk halaman depan.
func (s *ReportService) Dashboard(ctx context.Context) (*domain.DashboardSummary, error) {
	counters, err := s.reports.DashboardCounters(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	recentOrders, _, err := s.orders.List(ctx, s.pool,
		pagination.Params{Page: 1, PerPage: 8, Order: "desc"}, repository.OrderFilter{})
	if err != nil {
		return nil, err
	}

	upcomingTrips, _, err := s.trips.List(ctx, s.pool,
		pagination.Params{Page: 1, PerPage: 5, Sort: "depart_date", Order: "asc"}, domain.TripOpen)
	if err != nil {
		return nil, err
	}
	trips := make([]domain.Trip, 0, len(upcomingTrips))
	for _, t := range upcomingTrips {
		trips = append(trips, t.Trip)
	}

	topProducts, err := s.reports.ProductSales(ctx, s.pool, 5, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if topProducts == nil {
		topProducts = []domain.ProductSales{}
	}

	return &domain.DashboardSummary{
		ActiveTrips:     counters.ActiveTrips,
		OpenOrders:      counters.OpenOrders,
		PendingShipment: counters.PendingShipment,
		Outstanding:     money.RoundRupiah(counters.Outstanding),

		RevenueThisMonth: money.RoundRupiah(counters.RevenueThisMonth),
		ProfitThisMonth:  money.RoundRupiah(counters.RevenueThisMonth.Sub(counters.COGSThisMonth)),
		OrdersThisMonth:  counters.OrdersThisMonth,

		StockValue:    money.RoundRupiah(counters.StockValue),
		StockQty:      counters.StockQty,
		CustomerCount: counters.CustomerCount,

		RecentOrders:  recentOrders,
		UpcomingTrips: trips,
		TopProducts:   topProducts,
	}, nil
}
