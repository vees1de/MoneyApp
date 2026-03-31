package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const settingsManagePermissionCode = "settings.manage"

type Claims struct {
	UserID            string   `json:"uid"`
	SessionID         string   `json:"sid"`
	RoleCodes         []string `json:"roles,omitempty"`
	PermissionCodes   []string `json:"perms,omitempty"`
	EmployeeProfileID *string  `json:"epid,omitempty"`
	DepartmentID      *string  `json:"deptid,omitempty"`
	jwt.RegisteredClaims
}

type Principal struct {
	UserID            uuid.UUID
	SessionID         uuid.UUID
	RoleCodes         []string
	PermissionCodes   []string
	EmployeeProfileID *uuid.UUID
	DepartmentID      *uuid.UUID
}

type JWTManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewJWTManager(secret, issuer string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

func (m *JWTManager) SignAccessToken(userID, sessionID uuid.UUID, now time.Time) (string, error) {
	return m.SignPrincipalToken(Principal{
		UserID:    userID,
		SessionID: sessionID,
	}, now)
}

func (m *JWTManager) SignPrincipalToken(principal Principal, now time.Time) (string, error) {
	var employeeProfileID *string
	if principal.EmployeeProfileID != nil {
		value := principal.EmployeeProfileID.String()
		employeeProfileID = &value
	}

	var departmentID *string
	if principal.DepartmentID != nil {
		value := principal.DepartmentID.String()
		departmentID = &value
	}

	claims := Claims{
		UserID:            principal.UserID.String(),
		SessionID:         principal.SessionID.String(),
		RoleCodes:         principal.RoleCodes,
		PermissionCodes:   WithImplicitPermissions(principal.PermissionCodes),
		EmployeeProfileID: employeeProfileID,
		DepartmentID:      departmentID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   principal.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseAccessToken(token string) (*Principal, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(*jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil {
		return nil, err
	}

	var employeeProfileID *uuid.UUID
	if claims.EmployeeProfileID != nil {
		value, err := uuid.Parse(*claims.EmployeeProfileID)
		if err != nil {
			return nil, err
		}
		employeeProfileID = &value
	}

	var departmentID *uuid.UUID
	if claims.DepartmentID != nil {
		value, err := uuid.Parse(*claims.DepartmentID)
		if err != nil {
			return nil, err
		}
		departmentID = &value
	}

	return &Principal{
		UserID:            userID,
		SessionID:         sessionID,
		RoleCodes:         claims.RoleCodes,
		PermissionCodes:   WithImplicitPermissions(claims.PermissionCodes),
		EmployeeProfileID: employeeProfileID,
		DepartmentID:      departmentID,
	}, nil
}

type principalContextKey struct{}

func ContextWithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func WithImplicitPermissions(permissionCodes []string) []string {
	if hasPermissionCode(permissionCodes, settingsManagePermissionCode) {
		return permissionCodes
	}

	codes := append([]string{}, permissionCodes...)
	return append(codes, settingsManagePermissionCode)
}

func (p Principal) HasPermission(code string) bool {
	for _, item := range WithImplicitPermissions(p.PermissionCodes) {
		if item == code {
			return true
		}
	}

	return false
}

func hasPermissionCode(permissionCodes []string, target string) bool {
	for _, code := range permissionCodes {
		if code == target {
			return true
		}
	}

	return false
}
