package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	_ "embed"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//go:embed index.html
var indexHtml []byte

const interval = 10 // seconds between TOTP pins
const oldPins = 1   // also accept a TOTP pin from this many intervals ago
var cookieName string
var passHash string
var totpKey []byte
var sessionLifetime = 90 * time.Minute
var listenAddress string

var currentSession string
var sessionExpiry time.Time

func main() {
	initConfig()
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		session, _ := r.Cookie(cookieName)
		if session != nil && validSession(session.Value) {
			fmt.Fprint(w, "OK")
		} else {
			redirect := r.Header.Get("X-Forwarded-Uri")
			http.Redirect(w, r, "/login?r="+redirect, http.StatusSeeOther)
		}
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write(indexHtml)
		case http.MethodPost:
			r.ParseForm()
			if validLogin(r.PostForm.Get("pass"), r.PostForm.Get("totp")) {
				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    createSessionId(),
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				redirect := r.URL.Query().Get("r")
				http.Redirect(w, r, redirect, http.StatusSeeOther)
			} else {
				http.Error(w, "Invalid", http.StatusForbidden)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func initConfig() {
	cookieName = env("SPASM_COOKIE_NAME", "id")
	passHash = env("SPASM_PASS_HASH", "")
	var err error
	totpKey, err = base32.StdEncoding.DecodeString(env("SPASM_TOTP_KEY", ""))
	if err != nil {
		log.Fatalf("SPASM_TOTP_KEY is not valid base32: %s", err)
	}
	if len(totpKey) < sha512.Size {
		log.Printf("warning: SPASM_TOTP_KEY is %d bytes long, should be at least %d for maximum security", len(totpKey), sha512.Size)
	}
	listenAddress = env("SPASM_ADDRESS", "localhost:5000")
}

func validSession(session string) bool {
	return session == currentSession && sessionExpiry.After(time.Now())
}

func validLogin(pass, totp string) bool {
	return validPass(pass) && validTotp(totp)
}

func validPass(pass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passHash), []byte(pass)) == nil
}

func validTotp(totp string) bool {
	got64, err := strconv.ParseUint(totp, 10, 32)
	if err != nil {
		return false
	}
	got := uint32(got64)
	now := uint64(time.Now().Unix())
	for i := uint64(0); i <= oldPins; i++ {
		valid := genTotp(now - i*interval)
		if valid == got {
			return true
		}
	}
	return false
}

// refer to https://datatracker.ietf.org/doc/html/rfc4226
func genTotp(time uint64 /* unix timestamp */) uint32 {
	mac := hmac.New(sha512.New, totpKey)
	mac.Write(uint64ToBytes(time / interval))
	HS := mac.Sum([]byte{})
	if len(HS) != sha512.Size {
		log.Panic("invalid HMAC length")
	}
	Offset := HS[len(HS)-1] & 0xf
	Snum := binary.BigEndian.Uint32(HS[Offset:Offset+4]) &^ (1 << 31)
	return Snum % 1e6 // 6 digits
}

func createSessionId() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Panicf("cannot generate random value: %s", err)
	}
	currentSession = base64.URLEncoding.EncodeToString(bytes)
	sessionExpiry = time.Now().Add(sessionLifetime)
	log.Printf("Current session cookie %s expires at %v", currentSession, sessionExpiry)
	return currentSession
}

func uint64ToBytes(x uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, x)
	return b
}

func env(name, defaultVal string) string {
	v := os.Getenv(name)
	if v != "" {
		return v
	}
	if defaultVal != "" {
		return defaultVal
	}
	log.Fatalf("%s unset", name)
	return "" // will not get here
}
