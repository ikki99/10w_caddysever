package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Username  string
	ExpiresAt time.Time
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func CreateSession(username string) (string, error) {
	sessionID, err := GenerateSessionID()
	if err != nil {
		return "", err
	}

	mu.Lock()
	sessions[sessionID] = &Session{
		Username:  username,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mu.Unlock()

	return sessionID, nil
}

func GetSession(sessionID string) (*Session, bool) {
	mu.RLock()
	defer mu.RUnlock()

	session, exists := sessions[sessionID]
	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(sessions, sessionID)
		return nil, false
	}

	return session, true
}

func DeleteSession(sessionID string) {
	mu.Lock()
	delete(sessions, sessionID)
	mu.Unlock()
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		_, exists := GetSession(cookie.Value)
		if !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
