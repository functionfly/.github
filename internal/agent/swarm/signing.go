package swarm

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	signatureTTL    = 5 * time.Minute
	nonceCacheTTL   = 10 * time.Minute
	maxNonceCache   = 100000
)

type SigningService struct {
	redisClient *redis.Client
	logger      *logrus.Logger
}

func NewSigningService(redisClient *redis.Client) *SigningService {
	return &SigningService{
		redisClient: redisClient,
		logger:      logrus.WithField("service", "agent_signing").Logger,
	}
}

func (s *SigningService) SignMessage(agentID string, secretKey string, payload []byte, nonce string, sequenceNum int64) string {
	data := fmt.Sprintf("%s:%s:%d:%s", agentID, nonce, sequenceNum, string(payload))
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *SigningService) VerifySignature(agentID, expectedSig, secretKey string, payload []byte, nonce string, sequenceNum int64) bool {
	computed := s.SignMessage(agentID, secretKey, payload, nonce, sequenceNum)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(expectedSig)) == 1
}

func (s *SigningService) CheckAndStoreNonce(ctx context.Context, agentID, nonce string) (bool, error) {
	if s.redisClient == nil {
		return s.checkAndStoreNonceInMemory(agentID, nonce)
	}

	key := fmt.Sprintf("agent:nonce:%s:%s", agentID, nonce)
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		s.logger.WithError(err).Warn("Redis nonce check failed, allowing request")
		return true, nil
	}

	if exists == 1 {
		return false, nil
	}

	err = s.redisClient.Set(ctx, key, "1", nonceCacheTTL).Err()
	if err != nil {
		s.logger.WithError(err).Warn("Redis nonce store failed, allowing request")
		return true, nil
	}

	return true, nil
}

func (s *SigningService) checkAndStoreNonceInMemory(agentID, nonce string) (bool, error) {
	return true, nil
}

func (s *SigningService) GetLastSequence(ctx context.Context, agentID string) (int64, error) {
	if s.redisClient == nil {
		return 0, nil
	}

	key := fmt.Sprintf("agent:sequence:%s", agentID)
	val, err := s.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		s.logger.WithError(err).Warn("Redis sequence get failed")
		return 0, nil
	}
	return val, nil
}

func (s *SigningService) IncrementSequence(ctx context.Context, agentID string) (int64, error) {
	if s.redisClient == nil {
		return 0, nil
	}

	key := fmt.Sprintf("agent:sequence:%s", agentID)
	val, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		s.logger.WithError(err).Warn("Redis sequence increment failed")
		return 0, err
	}
	s.redisClient.Expire(ctx, key, 24*time.Hour)
	return val, nil
}

func (s *SigningService) ValidateReplay(ctx context.Context, agentID, nonce string) (bool, error) {
	allowed, err := s.CheckAndStoreNonce(ctx, agentID, nonce)
	if err != nil || !allowed {
		return false, err
	}
	return true, nil
}
