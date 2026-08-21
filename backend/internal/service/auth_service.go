// Package service berisi aturan bisnis aplikasi.
//
// Handler HTTP hanya menerjemahkan request menjadi pemanggilan service, dan
// repository hanya bicara SQL. Semua keputusan — boleh atau tidaknya sebuah
// operasi, bagaimana angka dihitung, kapan status berpindah — ada di sini.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/ipoool/jastipin/backend/internal/db"
	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/pagination"
	"github.com/ipoool/jastipin/backend/internal/pkg/token"
	"github.com/ipoool/jastipin/backend/internal/repository"
)

// bcryptCost 12 adalah kompromi wajar untuk aplikasi back office: cukup mahal
// untuk menahan brute force, tapi login tetap terasa instan.
const bcryptCost = 12

type AuthService struct {
	pool   *pgxpool.Pool
	users  *repository.UserRepo
	roles  *repository.RoleRepo
	tokens *token.Manager
}

func NewAuthService(pool *pgxpool.Pool, users *repository.UserRepo, roles *repository.RoleRepo, tokens *token.Manager) *AuthService {
	return &AuthService{pool: pool, users: users, roles: roles, tokens: tokens}
}

type Session struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         *domain.User `json:"user"`
}

// Login memverifikasi kredensial lalu menerbitkan sepasang token.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*Session, error) {
	// Dinormalkan sekali di sini supaya hitungan kegagalan tidak bisa diakali
	// dengan mengubah huruf besar-kecil emailnya.
	email = strings.ToLower(strings.TrimSpace(email))
	now := time.Now()

	attempt, err := s.users.GetLoginAttempt(ctx, s.pool, email)
	if err != nil {
		return nil, err
	}
	if sisa := attempt.BlockedFor(now); sisa > 0 {
		return nil, domain.TooMany(
			"terlalu banyak percobaan login yang gagal — coba lagi dalam %s", sisaMenit(sisa))
	}

	user, err := s.users.GetByEmail(ctx, s.pool, email)
	if err != nil {
		if domainErr, ok := domain.AsError(err); ok && domainErr.Code == domain.CodeNotFound {
			// Jangan bocorkan apakah email terdaftar: pesan yang sama dipakai
			// untuk email tak dikenal maupun password salah, dan kegagalannya
			// sama-sama dihitung.
			return nil, s.catatGagalLogin(ctx, email, ipAddress)
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, s.catatGagalLogin(ctx, email, ipAddress)
	}

	// Akun nonaktif tidak ikut dihitung sebagai percobaan gagal: passwordnya
	// benar, dan menguncinya hanya menyusahkan orang yang memang berhak tahu
	// kenapa ia tidak bisa masuk.
	if !user.IsActive {
		return nil, domain.Forbidden("akun ini sudah dinonaktifkan, hubungi owner")
	}

	session, err := s.issueSession(ctx, user, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	// Kegagalan membersihkan rekaman atau mencatat waktu login tidak boleh
	// menggagalkan login yang sudah sah.
	_ = s.users.ClearLoginAttempts(ctx, s.pool, email)
	_ = s.users.TouchLastLogin(ctx, s.pool, user.ID)
	return session, nil
}

// catatGagalLogin menaikkan hitungan kegagalan lalu menyusun pesan penolakannya.
//
// Sisa percobaan ikut disebutkan supaya pengguna yang cuma salah ketik tahu ia
// mendekati batas. Itu tidak membocorkan apa pun: hitungannya berlaku sama untuk
// email yang terdaftar maupun tidak.
func (s *AuthService) catatGagalLogin(ctx context.Context, email, ipAddress string) error {
	attempt, err := s.users.RecordFailedLogin(
		ctx, s.pool, email, ipAddress, domain.LoginMaxAttempts, domain.LoginBlockDuration)
	if err != nil {
		// Penjagaannya gagal, tapi passwordnya memang salah. Menolak dengan
		// pesan biasa lebih benar daripada membalas galat server.
		return domain.Unauthorized("email atau password salah")
	}

	// Membersihkan baris lama sekalian, selagi sudah menyentuh tabelnya.
	_ = s.users.PurgeStaleLoginAttempts(ctx, s.pool, 24*time.Hour)

	if sisa := attempt.BlockedFor(time.Now()); sisa > 0 {
		return domain.TooMany(
			"gagal %d kali berturut-turut — login untuk email ini dikunci %s",
			domain.LoginMaxAttempts, sisaMenit(sisa))
	}

	tersisa := attempt.AttemptsLeft(time.Now())
	if tersisa <= 2 {
		return domain.Unauthorized(fmt.Sprintf(
			"email atau password salah — sisa %d percobaan sebelum dikunci %s",
			tersisa, sisaMenit(domain.LoginBlockDuration)))
	}
	return domain.Unauthorized("email atau password salah")
}

// sisaMenit membulatkan durasi ke atas ke menit penuh.
//
// Detik yang tercetak apa adanya ("4m37s") hanya membuat orang menghitung
// sendiri; yang mereka butuhkan cuma tahu kira-kira berapa lama menunggu.
//
// Pembulatannya dikerjakan sekali dengan pembagian ke atas. Versi sebelumnya
// membulatkan dulu lalu memeriksa sisa dari nilai aslinya, sehingga kunci lima
// menit dilaporkan sebagai enam.
func sisaMenit(d time.Duration) string {
	if d <= 0 {
		return "kurang dari 1 menit"
	}
	menit := int((d + time.Minute - 1) / time.Minute)
	if menit <= 1 {
		return "1 menit"
	}
	return fmt.Sprintf("%d menit", menit)
}

// Refresh menukar refresh token dengan sepasang token baru.
//
// Token lama langsung dicabut (rotasi): kalau refresh token bocor dan dipakai
// penyerang, pemakaian berikutnya oleh pemilik asli akan gagal, sehingga
// pembobolan cepat ketahuan.
func (s *AuthService) Refresh(ctx context.Context, rawToken, userAgent, ipAddress string) (*Session, error) {
	hashed := token.HashRefreshToken(rawToken)

	stored, err := s.users.GetRefreshToken(ctx, s.pool, hashed)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Unauthorized("sesi tidak ditemukan, silakan login ulang")
		}
		if domainErr, ok := domain.AsError(err); ok && domainErr.Code == domain.CodeNotFound {
			return nil, domain.Unauthorized("sesi tidak ditemukan, silakan login ulang")
		}
		return nil, err
	}
	if !stored.IsUsable() {
		return nil, domain.Unauthorized("sesi sudah berakhir, silakan login ulang")
	}

	user, err := s.users.GetByID(ctx, s.pool, stored.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, domain.Forbidden("akun ini sudah dinonaktifkan, hubungi owner")
	}

	var session *Session
	err = db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.RevokeRefreshToken(ctx, tx, hashed); err != nil {
			return err
		}
		session, err = s.issueSessionTx(ctx, tx, user, userAgent, ipAddress)
		return err
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// Logout mencabut satu sesi.
func (s *AuthService) Logout(ctx context.Context, rawToken string) error {
	return s.users.RevokeRefreshToken(ctx, s.pool, token.HashRefreshToken(rawToken))
}

// LogoutAll mencabut seluruh sesi pengguna, dipakai saat password diganti.
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.users.RevokeAllRefreshTokens(ctx, s.pool, userID)
}

// Me mengambil identitas pemilik token yang sedang dipakai.
//
// Akun yang sudah tidak ada atau dinonaktifkan dijawab UNAUTHORIZED, bukan
// NOT_FOUND. Bedanya menentukan pengalaman pengguna: NOT_FOUND membuat aplikasi
// mengira ada data yang hilang lalu menampilkan galat 404 di setiap halaman,
// sedangkan UNAUTHORIZED adalah kenyataannya — sesinya tidak berlaku lagi —
// sehingga proxy bisa mencoba memperbarui token dan, kalau gagal, mengantar
// pengguna ke halaman login dengan bersih.
func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, s.pool, userID)
	if err != nil {
		var domainErr *domain.Error
		if errors.As(err, &domainErr) && domainErr.Code == domain.CodeNotFound {
			return nil, domain.Unauthorized("sesi tidak berlaku lagi, silakan login ulang")
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, domain.Unauthorized("akun ini sedang dinonaktifkan")
	}

	// Hak akses efektif ikut dikirim: sidebar dan penjaga rute di frontend
	// membacanya dari sini. Tanpa itu, seluruh menu tersaring habis dan yang
	// terlihat adalah sidebar kosong melompong.
	role, err := s.roles.Get(ctx, s.pool, user.Role)
	if err != nil {
		return nil, err
	}
	return withRole(user, *role), nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.users.GetByID(ctx, s.pool, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return domain.Validation("password saat ini salah", map[string]string{
			"current_password": "tidak cocok dengan password akun",
		})
	}

	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.UpdatePassword(ctx, tx, userID, hashed); err != nil {
			return err
		}
		// Semua perangkat lain dipaksa login ulang dengan password baru.
		return s.users.RevokeAllRefreshTokens(ctx, tx, userID)
	})
}

func (s *AuthService) issueSession(ctx context.Context, user *domain.User, userAgent, ipAddress string) (*Session, error) {
	return s.issueSessionTx(ctx, s.pool, user, userAgent, ipAddress)
}

func (s *AuthService) issueSessionTx(ctx context.Context, q db.Querier, user *domain.User, userAgent, ipAddress string) (*Session, error) {
	// Wewenang dan daftar menu ikut ditanam di dalam token supaya middleware
	// tidak perlu menyentuh database tiap request. Konsekuensinya perubahan
	// role baru terasa saat token berikutnya terbit — karena itu RoleService
	// dan UserService mencabut sesi begitu haknya berubah.
	role, err := s.roles.Get(ctx, q, user.Role)
	if err != nil {
		return nil, err
	}
	withRole(user, *role)

	accessToken, expiresAt, err := s.tokens.IssueAccessToken(
		user.ID, user.Email, user.Role, role.Scope, user.EffectivePermissions)
	if err != nil {
		return nil, domain.Internal(err)
	}

	rawRefresh, hashedRefresh, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, domain.Internal(err)
	}

	if err := s.users.CreateRefreshToken(ctx, q, user.ID, hashedRefresh,
		time.Now().Add(s.tokens.RefreshTTL()), optionalString(userAgent), optionalString(ipAddress)); err != nil {
		return nil, err
	}

	return &Session{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}

// --- Manajemen pengguna ----------------------------------------------------

type UserService struct {
	pool  *pgxpool.Pool
	users *repository.UserRepo
	roles *repository.RoleRepo
}

func NewUserService(pool *pgxpool.Pool, users *repository.UserRepo, roles *repository.RoleRepo) *UserService {
	return &UserService{pool: pool, users: users, roles: roles}
}

// withRole mengisi label role dan hak akses efektif sebuah pengguna.
//
// Perhitungannya dikerjakan di service, bukan di handler, karena daftar menu
// sebuah role sekarang tinggal di database — dan handler tidak punya jalan ke
// sana. Antarmuka cukup membaca hasilnya; kalau tidak, aturan penggabungan yang
// sama harus ditulis ulang di frontend dan cepat atau lambat keduanya berbeda.
func withRole(user *domain.User, role domain.Role) *domain.User {
	if user == nil {
		return nil
	}
	user.RoleLabel = role.Label
	user.EffectivePermissions = domain.EffectivePermissions(role.Name, role.Permissions, user.Permissions)
	return user
}

type CreateUserInput struct {
	Name        string
	Email       string
	Password    string
	Role        string
	Phone       *string
	Permissions []string
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput, pemintaRole string) (*domain.User, error) {
	role, err := s.requireRole(ctx, in.Role, pemintaRole)
	if err != nil {
		return nil, err
	}

	hashed, err := HashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	permissions, err := sanitizePermissions(*role, in.Permissions)
	if err != nil {
		return nil, err
	}

	user, err := s.users.Create(ctx, s.pool, repository.CreateUserParams{
		Name:         strings.TrimSpace(in.Name),
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash: hashed,
		Role:         role.Name,
		Phone:        in.Phone,
		Permissions:  permissions,
	})
	if err != nil {
		return nil, err
	}
	return withRole(user, *role), nil
}

// jagaAkunRoot menolak sentuhan pada akun root dari siapa pun selain root.
//
// Dijawab "tidak ditemukan", bukan "tidak boleh": akun root memang tidak
// tampil di daftar, dan penolakan yang berbeda bunyinya justru memberi tahu
// bahwa ia ada.
func jagaAkunRoot(targetRole, pemintaRole string) error {
	if domain.IsRootRole(targetRole) && !domain.IsRootRole(pemintaRole) {
		return domain.NotFound("pengguna")
	}
	return nil
}

// requireRole memastikan role yang diminta benar-benar ada.
//
// Dulu daftarnya tertutup di kode dan cukup dicocokkan dengan tiga nama; kini
// role adalah data, jadi yang menentukan sah atau tidaknya adalah isi tabelnya.
func (s *UserService) requireRole(ctx context.Context, name, pemintaRole string) (*domain.Role, error) {
	name = strings.TrimSpace(name)

	// Role root tidak tampil di daftar pilihan, jadi permintaan yang menyebutnya
	// datang dari luar antarmuka. Melolosinya berarti siapa pun yang memegang
	// menu Pengguna bisa mengangkat dirinya sendiri jadi root.
	if domain.IsRootRole(name) && !domain.IsRootRole(pemintaRole) {
		return nil, domain.Validation("role tidak dikenal", map[string]string{
			"role": "pilih role yang tersedia di daftar",
		})
	}

	role, err := s.roles.Get(ctx, s.pool, name)
	if err != nil {
		if domainErr, ok := domain.AsError(err); ok && domainErr.Code == domain.CodeNotFound {
			return nil, domain.Validation("role tidak dikenal", map[string]string{
				"role": "pilih role yang tersedia di daftar",
			})
		}
		return nil, err
	}
	return role, nil
}

// List menyembunyikan akun root dari siapa pun selain root sendiri.
func (s *UserService) List(ctx context.Context, p pagination.Params, pemintaRole string) ([]domain.User, int64, error) {
	users, total, err := s.users.List(ctx, s.pool, p, domain.IsRootRole(pemintaRole))
	if err != nil {
		return nil, 0, err
	}

	// Seluruh role diambil sekali, bukan satu query per baris untuk daftar menu
	// yang itu-itu juga.
	roles, err := s.roles.Map(ctx, s.pool)
	if err != nil {
		return nil, 0, err
	}
	for i := range users {
		withRole(&users[i], roles[users[i].Role])
	}
	return users, total, nil
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID, pemintaRole string) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	if err := jagaAkunRoot(user.Role, pemintaRole); err != nil {
		return nil, err
	}
	role, err := s.roles.Get(ctx, s.pool, user.Role)
	if err != nil {
		return nil, err
	}
	return withRole(user, *role), nil
}

type UpdateUserInput struct {
	Name     string
	Role     string
	Phone    *string
	IsActive bool
	// Permissions kosong berarti kembali mengikuti bawaan role.
	Permissions []string
}

// sanitizePermissions membuang hak akses yang tidak dikenal dan yang melampaui
// batas role.
//
// Batas role adalah keputusan keamanan; centang di antarmuka hanya boleh
// mempersempitnya, tidak pernah melebarkan. Tanpa penyaringan ini, seorang
// tripper bisa diberi menu pengaturan hanya dengan mengirim permintaan yang
// dirakit sendiri.
func sanitizePermissions(role domain.Role, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return nil, nil
	}

	for _, p := range requested {
		if !domain.IsValidPermission(p) {
			return nil, domain.Validation("hak akses tidak dikenal", map[string]string{
				"permissions": p + " bukan menu yang ada di aplikasi",
			})
		}
	}

	effective := domain.EffectivePermissions(role.Name, role.Permissions, requested)
	if len(effective) == 0 {
		return nil, domain.Validation("hak akses tidak sesuai role", map[string]string{
			"permissions": "tidak ada satu pun menu yang boleh dibuka role ini",
		})
	}
	return effective, nil
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput, pemintaRole string) (*domain.User, error) {
	role, err := s.requireRole(ctx, in.Role, pemintaRole)
	if err != nil {
		return nil, err
	}

	current, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return nil, err
	}
	if err := jagaAkunRoot(current.Role, pemintaRole); err != nil {
		return nil, err
	}

	permissions, err := sanitizePermissions(*role, in.Permissions)
	if err != nil {
		return nil, err
	}

	// Akun terakhir yang masih bisa membuka menu Pengguna tidak boleh
	// kehilangan menunya atau dinonaktifkan: menu itu satu-satunya jalan
	// mengembalikan hak akses siapa pun, dan sekali hilang pemulihannya cuma
	// lewat database.
	//
	// Yang diperiksa daftar menunya, bukan nama rolenya. Sejak role jadi data,
	// "yang bisa mengelola pengguna" tidak lagi berarti "yang bernama owner".
	tetapMengelola := in.IsActive && domain.HasPermission(
		domain.EffectivePermissions(role.Name, role.Permissions, permissions), domain.PermUsers)
	if !tetapMengelola {
		remaining, err := s.users.CountActiveUserManagers(ctx, s.pool, id)
		if err != nil {
			return nil, err
		}
		if remaining == 0 {
			return nil, domain.Conflict(
				"ini akun terakhir yang bisa mengelola pengguna — beri akses itu ke akun lain terlebih dahulu")
		}
	}

	updated, err := s.users.Update(ctx, s.pool, id, repository.UpdateUserParams{
		Name:        strings.TrimSpace(in.Name),
		Role:        role.Name,
		Phone:       in.Phone,
		IsActive:    in.IsActive,
		Permissions: permissions,
	})
	if err != nil {
		return nil, err
	}

	// Hak akses ikut dibawa di dalam access token, jadi perubahannya tidak akan
	// terasa sampai token berikutnya terbit. Sesi pengguna itu dicabut supaya
	// pembatasan berlaku saat itu juga — termasuk saat hak akses dicabut karena
	// alasan mendesak.
	if permissionsChanged(current, updated) {
		if err := s.users.RevokeAllRefreshTokens(ctx, s.pool, id); err != nil {
			return nil, err
		}
	}
	return withRole(updated, *role), nil
}

func permissionsChanged(before, after *domain.User) bool {
	if before.Role != after.Role || len(before.Permissions) != len(after.Permissions) {
		return true
	}
	for i := range before.Permissions {
		if before.Permissions[i] != after.Permissions[i] {
			return true
		}
	}
	return false
}

// ResetPassword dipakai owner untuk mengganti password pengguna lain yang lupa.
func (s *UserService) ResetPassword(ctx context.Context, id uuid.UUID, newPassword, pemintaRole string) error {
	target, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return err
	}
	if err := jagaAkunRoot(target.Role, pemintaRole); err != nil {
		return err
	}

	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.users.UpdatePassword(ctx, tx, id, hashed); err != nil {
			return err
		}
		return s.users.RevokeAllRefreshTokens(ctx, tx, id)
	})
}

func (s *UserService) Delete(ctx context.Context, id, actorID uuid.UUID, pemintaRole string) error {
	if id == actorID {
		return domain.Conflict("tidak bisa menghapus akun yang sedang kamu pakai")
	}

	user, err := s.users.GetByID(ctx, s.pool, id)
	if err != nil {
		return err
	}
	if err := jagaAkunRoot(user.Role, pemintaRole); err != nil {
		return err
	}
	role, err := s.roles.Get(ctx, s.pool, user.Role)
	if err != nil {
		return err
	}

	// Sama seperti penurunan role: yang dijaga bukan nama rolenya, melainkan
	// supaya selalu tersisa satu akun yang bisa membuka menu Pengguna.
	if domain.HasPermission(
		domain.EffectivePermissions(role.Name, role.Permissions, user.Permissions), domain.PermUsers) {
		remaining, err := s.users.CountActiveUserManagers(ctx, s.pool, id)
		if err != nil {
			return err
		}
		if remaining == 0 {
			return domain.Conflict("ini akun terakhir yang bisa mengelola pengguna")
		}
	}

	return s.users.Delete(ctx, s.pool, id)
}

// HashPassword membuat hash bcrypt. Dipakai bersama oleh service dan perintah seed.
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", domain.Validation("password terlalu pendek", map[string]string{
			"password": "minimal 8 karakter",
		})
	}
	// bcrypt memotong input di 72 byte; menolaknya lebih jujur daripada
	// diam-diam mengabaikan sisa karakter yang diketik pengguna.
	if len(password) > 72 {
		return "", domain.Validation("password terlalu panjang", map[string]string{
			"password": "maksimal 72 karakter",
		})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", domain.Internal(err)
	}
	return string(hashed), nil
}

func optionalString(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}
