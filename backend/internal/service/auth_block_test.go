package service

import (
	"testing"
	"time"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

func TestSisaMenit(t *testing.T) {
	tests := []struct {
		nama string
		d    time.Duration
		mau  string
	}{
		// Kunci lima menit dilaporkan lima menit, bukan enam. Sisa waktunya
		// selalu sedikit di bawah lima menit saat pesannya disusun.
		{"tepat lima menit", 5 * time.Minute, "5 menit"},
		{"sekejap di bawah lima menit", 5*time.Minute - time.Millisecond, "5 menit"},
		{"empat menit lewat sedikit", 4*time.Minute + time.Second, "5 menit"},
		{"tepat satu menit", time.Minute, "1 menit"},
		{"beberapa detik", 3 * time.Second, "1 menit"},
		{"habis", 0, "kurang dari 1 menit"},
	}
	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			if got := sisaMenit(tt.d); got != tt.mau {
				t.Errorf("sisaMenit(%v) = %q, mau %q", tt.d, got, tt.mau)
			}
		})
	}
}

func TestLoginAttemptSisaPercobaan(t *testing.T) {
	now := time.Now()
	baru := func(count int, lalu time.Duration) *domain.LoginAttempt {
		t := now.Add(-lalu)
		return &domain.LoginAttempt{FailedCount: count, LastFailedAt: &t}
	}

	if got := (*domain.LoginAttempt)(nil).AttemptsLeft(now); got != domain.LoginMaxAttempts {
		t.Errorf("tanpa rekaman seharusnya %d percobaan, dapat %d", domain.LoginMaxAttempts, got)
	}
	if got := baru(3, time.Minute).AttemptsLeft(now); got != 2 {
		t.Errorf("3 gagal dalam jendela seharusnya sisa 2, dapat %d", got)
	}
	// Kegagalan lama tidak boleh ikut menghitung: satu salah ketik bulan lalu
	// tidak boleh mendekatkan siapa pun ke penguncian hari ini.
	if got := baru(4, 2*domain.LoginBlockDuration).AttemptsLeft(now); got != domain.LoginMaxAttempts {
		t.Errorf("kegagalan di luar jendela seharusnya dilupakan, dapat sisa %d", got)
	}
}

func TestLoginAttemptBlockedFor(t *testing.T) {
	now := time.Now()
	nanti := now.Add(3 * time.Minute)
	lalu := now.Add(-time.Minute)

	if got := (&domain.LoginAttempt{BlockedUntil: &nanti}).BlockedFor(now); got <= 0 {
		t.Error("kunci yang belum lewat seharusnya masih menahan")
	}
	if got := (&domain.LoginAttempt{BlockedUntil: &lalu}).BlockedFor(now); got != 0 {
		t.Errorf("kunci yang sudah lewat seharusnya lepas, dapat %v", got)
	}
	if got := (*domain.LoginAttempt)(nil).BlockedFor(now); got != 0 {
		t.Errorf("tanpa rekaman seharusnya tidak terkunci, dapat %v", got)
	}
}
