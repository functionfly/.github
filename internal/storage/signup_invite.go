package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	// ErrSignupInviteInvalid is returned when the code does not match or the invite row is missing.
	ErrSignupInviteInvalid = errors.New("invalid invite code")
	// ErrSignupInviteExhausted is returned when max_uses has been reached.
	ErrSignupInviteExhausted = errors.New("invite code has no uses remaining")
	// ErrSignupInviteRevoked is returned when the invite was revoked.
	ErrSignupInviteRevoked = errors.New("invite code has been revoked")
)

const signupInviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// SignupInviteNormalize trims and uppercases invite input for case-insensitive matching.
func SignupInviteNormalize(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

func signupInviteFingerprint(normalized string) string {
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// GenerateSignupInvitePlainCode returns a random human-enterable code (default length 16).
func GenerateSignupInvitePlainCode(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	n := len(signupInviteAlphabet)
	buf := make([]byte, length)
	max := 256 - (256 % n)
	for i := 0; i < length; {
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		if int(b[0]) >= max {
			continue
		}
		buf[i] = signupInviteAlphabet[int(b[0])%n]
		i++
	}
	return string(buf), nil
}

// CreateSignupInvite inserts a new invite; returns the row id and plaintext code (show once to admin).
func (db *PostgresDB) CreateSignupInvite(ctx context.Context, label string, maxUses *int, expiresAt *time.Time, createdBy *uuid.UUID) (id uuid.UUID, plainCode string, err error) {
	plainCode, err = GenerateSignupInvitePlainCode(16)
	if err != nil {
		return uuid.Nil, "", err
	}
	norm := SignupInviteNormalize(plainCode)
	fp := signupInviteFingerprint(norm)
	hash, err := bcrypt.GenerateFromPassword([]byte(norm), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, "", err
	}
	row := SignupInviteCode{
		ID:              uuid.New(),
		CodeFingerprint: fp,
		CodeHash:        string(hash),
		Label:           strings.TrimSpace(label),
		MaxUses:         maxUses,
		UsesCount:       0,
		ExpiresAt:       expiresAt,
		CreatedBy:       createdBy,
	}
	if err := db.GORM.WithContext(ctx).Create(&row).Error; err != nil {
		return uuid.Nil, "", err
	}
	return row.ID, plainCode, nil
}

// ListSignupInvitesAdmin returns invite metadata for admin UI (no secrets).
func (db *PostgresDB) ListSignupInvitesAdmin(ctx context.Context) ([]SignupInviteCodeAdminList, error) {
	var rows []SignupInviteCode
	if err := db.GORM.WithContext(ctx).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]SignupInviteCodeAdminList, 0, len(rows))
	for _, r := range rows {
		out = append(out, SignupInviteCodeAdminList{
			ID:        r.ID,
			Label:     r.Label,
			MaxUses:   r.MaxUses,
			UsesCount: r.UsesCount,
			ExpiresAt: r.ExpiresAt,
			RevokedAt: r.RevokedAt,
			CreatedBy: r.CreatedBy,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}

// RevokeSignupInvite sets revoked_at on an invite.
func (db *PostgresDB) RevokeSignupInvite(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	res := db.GORM.WithContext(ctx).Model(&SignupInviteCode{}).Where("id = ? AND revoked_at IS NULL", id).Update("revoked_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ValidateSignupInviteReadOnly checks code without consuming a use (e.g. OAuth URL preflight).
func (db *PostgresDB) ValidateSignupInviteReadOnly(ctx context.Context, plainCode string) error {
	norm := SignupInviteNormalize(plainCode)
	if norm == "" {
		return ErrSignupInviteInvalid
	}
	fp := signupInviteFingerprint(norm)
	var inv SignupInviteCode
	err := db.GORM.WithContext(ctx).
		Where("code_fingerprint = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?)", fp, time.Now().UTC()).
		First(&inv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSignupInviteInvalid
		}
		return err
	}
	if inv.RevokedAt != nil {
		return ErrSignupInviteRevoked
	}
	if inv.MaxUses != nil && inv.UsesCount >= *inv.MaxUses {
		return ErrSignupInviteExhausted
	}
	if err := bcrypt.CompareHashAndPassword([]byte(inv.CodeHash), []byte(norm)); err != nil {
		return ErrSignupInviteInvalid
	}
	return nil
}

// ReserveSignupInvite increments uses_count after validating the code (call ReleaseSignupInviteReservation on signup failure).
func (db *PostgresDB) ReserveSignupInvite(ctx context.Context, plainCode string) (inviteID uuid.UUID, err error) {
	norm := SignupInviteNormalize(plainCode)
	if norm == "" {
		return uuid.Nil, ErrSignupInviteInvalid
	}
	fp := signupInviteFingerprint(norm)
	err = db.GORM.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv SignupInviteCode
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code_fingerprint = ? AND revoked_at IS NULL AND (expires_at IS NULL OR expires_at > ?)", fp, time.Now().UTC())
		if err := q.First(&inv).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSignupInviteInvalid
			}
			return err
		}
		if inv.MaxUses != nil && inv.UsesCount >= *inv.MaxUses {
			return ErrSignupInviteExhausted
		}
		if err := bcrypt.CompareHashAndPassword([]byte(inv.CodeHash), []byte(norm)); err != nil {
			return ErrSignupInviteInvalid
		}
		newCount := inv.UsesCount + 1
		res := tx.Model(&SignupInviteCode{}).Where("id = ? AND uses_count = ?", inv.ID, inv.UsesCount).Updates(map[string]interface{}{
			"uses_count": newCount,
			"updated_at": time.Now().UTC(),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrSignupInviteExhausted
		}
		inviteID = inv.ID
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return inviteID, nil
}

// ReleaseSignupInviteReservation decrements uses_count after a failed signup following ReserveSignupInvite.
func (db *PostgresDB) ReleaseSignupInviteReservation(ctx context.Context, inviteID uuid.UUID) error {
	res := db.GORM.WithContext(ctx).Model(&SignupInviteCode{}).
		Where("id = ? AND uses_count > 0", inviteID).
		Updates(map[string]interface{}{
			"uses_count": gorm.Expr("uses_count - 1"),
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
