package jwtx

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
	ErrRevokeToken  = errors.New("token revoked")
)

type jwtManager struct {
	sectectKey  []byte
	iss         string
	aud         string
	exp         time.Duration
	leeway      time.Duration
	redisClient *redis.Client
	prefix      string
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Options struct {
	Secret      string        // HMAC secret
	Issuer      string        // iss
	Audience    string        // aud (optional)
	Expiry      time.Duration // exp (24h)
	Leeway      time.Duration // clock skew allowance (30sec)
	RedisPrefix string        // prefix key blacklist เช่น "jwt:blacklist:"
}

type TokenManager interface {
	GenerateToken(userID uint, username string, role string) (token string, jwtID string, err error)
	ValidateToken(token string) (*Claims, error)
	BlacklistToken(ctx context.Context, jwtID string, exp time.Duration) error
	IsBlacklisted(ctx context.Context, jwtID string) (bool, error)
	GetExpiry(token string) (time.Duration, error)
}

func NewJWTManager(opt Options, redisClient *redis.Client) TokenManager {
	pfx := opt.RedisPrefix
	if pfx == "" {
		pfx = "jwt:blacklist:"
	}
	return &jwtManager{
		sectectKey:  []byte(opt.Secret),
		iss:         opt.Issuer,
		aud:         opt.Audience,
		exp:         opt.Expiry,
		leeway:      opt.Leeway,
		redisClient: redisClient,
		prefix:      pfx,
	}
}

// Generate สร้าง Token พร้อม JWTID แบบ crypto-safe
func (j *jwtManager) GenerateToken(userID uint, username, role string) (string, string, error) {
	now := time.Now().UTC()
	jwtID, err := randomJTI(16)
	if err != nil {
		return "", "", err
	}

	rc := jwt.RegisteredClaims{
		ID:        jwtID,
		Issuer:    j.iss,
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now.Add(-j.leeway)), // กัน clock skew ฝัง client
		ExpiresAt: jwt.NewNumericDate(now.Add(j.exp)),
	}

	if j.aud != "" {
		rc.Audience = []string{j.aud}
	}

	claims := &Claims{
		UserID:           userID,
		Username:         username,
		Role:             role,
		RegisteredClaims: rc,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(j.sectectKey)
	return signed, jwtID, err
}

func (j *jwtManager) ValidateToken(token string) (*Claims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrInvalidToken
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Name}),
		jwt.WithLeeway(j.leeway),
	}
	if j.iss != "" {
		opts = append(opts, jwt.WithIssuer(j.iss))
	}
	if j.aud != "" {
		opts = append(opts, jwt.WithAudience(j.aud))
	}

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return j.sectectKey, nil
	}, opts...)
	if err != nil {
		// แยกระหว่าง token หมดอายุ กับ Error อื่นๆ
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (j *jwtManager) BlacklistToken(ctx context.Context, jwtID string, exp time.Duration) error {
	if j.redisClient == nil || exp <= 0 {
		return nil // ไม่เปิดใช้ blacklist
	}
	return j.redisClient.Set(ctx, j.prefix+jwtID, "1", exp).Err()
}

func (j *jwtManager) IsBlacklisted(ctx context.Context, jwtID string) (bool, error) {
	if j.redisClient == nil || jwtID == "" {
		return false, nil
	}
	ok, err := j.redisClient.Exists(ctx, j.prefix+jwtID).Result()
	return ok == 1, err
}

// GetExpiry คืนค่าเวลาที่เหลือก่อนหมดอายุของ Token (ใช้ตอน logout ถ้าอยาก blacklist พอดี exp)
func (j *jwtManager) GetExpiry(token string) (time.Duration, error) {
	Claims, err := j.ValidateToken(token)
	if err != nil {
		return 0, err
	}
	if Claims.ExpiresAt == nil {
		return 0, errors.New("no exp")
	}
	d := time.Until(Claims.ExpiresAt.Time)
	if d < 0 {
		return 0, ErrExpiredToken
	}
	return d, nil
}

// helper function

// generate JWTID
func randomJTI(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// ใช้ base64 URL-safe ตัด "=" ออก
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(b), "="), nil
}

// ExtractBearer ดึง Token จาก header Autherization: Bearer <token>
func ExtractBearer(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// jwt.ParseWithClaims นี่คือเมธอดหลักที่รับผิดชอบการแยกส่วนและตรวจสอบ Token โดยจะรับ 3 พารามิเตอร์:
// tokenString: ค่าของ JWT Token ที่เป็น String ที่ได้รับจาก HTTP Request
// &Claims{}: Instance ของ Claims struct ที่ใช้เก็บข้อมูล Payload ที่ถูกถอดรหัส
// keyFunc: ฟังก์ชันที่ใช้หา Secret Key สำหรับการตรวจสอบ Signature

// เมื่อเมธอดนี้ถูกเรียกใช้ จะเกิดการทำงานเบื้องหลังดังนี้:
// แยก Token: โค้ดจะแยก Token ออกเป็น 3 ส่วนด้วยจุด . คือ Header, Payload, และ Signature
// Base64 Decode: ทำการถอดรหัส (Decode) ส่วน Header และ Payload จาก Base64 URL Safe Encoding
// ตรวจสอบ Signature: สร้าง Signature ใหม่จาก Header และ Payload ที่ถอดรหัสได้ แล้วนำมาเปรียบเทียบกับ Signature ที่ส่งมาใน Token หากไม่ตรงกันจะคืนค่า Error ทันที
