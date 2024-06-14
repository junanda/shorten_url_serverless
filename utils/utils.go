package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"strings"
	"time"

	// "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/junanda/shortenerUrl/pkg/model"
	"golang.org/x/crypto/bcrypt"
)

// ShortenURL takes a URL string and returns a shortened version of it.
func ShortenURL(url string) string {
	hasher := sha256.New()
	hasher.Write([]byte(url))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	// We use the first 8 characters for the shortened URL, which should be sufficient for non-cryptographic use
	return sha[:8]
}

func GenerateUUID() string {
	itemUUID := uuid.New().String()
	splitUId := strings.Split(itemUUID, "-")
	return strings.Join(splitUId, "")
}

// EncryptPassword takes a plain text password and returns a bcrypt hashed password.
func EncryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CompareHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func PrintError(judul string, err error) {
	log.Printf("%s : %s", judul, err.Error())
}

func GenerateToken(extUser model.User, secretjwt []byte) (string, error) {
	expirateTime := time.Now().Add(time.Hour * 1).Unix()

	claim := &model.Claims{
		Role: extUser.Role,
		Uid:  extUser.IdUser,
		StandardClaims: jwt.StandardClaims{
			Subject:   extUser.Username,
			ExpiresAt: expirateTime,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString(secretjwt)
	return tokenString, err
}

func ParseToken(tokenString string) (claim *model.Claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	claim, ok := token.Claims.(*model.Claims)
	if !ok {
		PrintError("Error claims extract", err)
		return claim, err
	}

	if err != nil {
		PrintError("Error parsing with claim", err)
		return claim, err
	}

	return claim, nil
}

func GenerateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 12
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SendEmail(email, subject, message string) error {

	emailMessage := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", email, subject, message)

	// Set up the SMTP client
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("EMAIL_SENDER")
	emailPass := os.Getenv("EMAIL_PASSWORD")
	auth := smtp.PlainAuth("", from, emailPass, smtpHost)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{email}, []byte(emailMessage))
	if err != nil {
		PrintError("Failed to send email", err)
		return err
	}

	return nil
}
