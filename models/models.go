package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/fiam/gounidecode/unidecode"
	"github.com/jinzhu/gorm"

	//postgres dialect, required by gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func init() {
	assertAvailablePRNG()
}

func assertAvailablePRNG() {
	buf := make([]byte, 1)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

//SetDB establishes connection to database and saves its handler into db *sqlx.DB
func SetDB(connection string) {
	var err error
	db, err = gorm.Open("postgres", connection)
	if err != nil {
		panic(err)
	}
}

//GetDB returns database handler
func GetDB() *gorm.DB {
	return db
}

//AutoMigrate runs gorm auto migration
func AutoMigrate() {
	db.AutoMigrate(&User{}, &Page{}, &MenuItem{}, &Post{}, &Slide{}, &Setting{})
}

//truncate truncates string to n runes
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

//createSlug makes url slug out of string
func createSlug(s string) string {
	s = strings.ToLower(unidecode.Unidecode(s))                     //transliterate if it is not in english
	s = regexp.MustCompile("[^a-z0-9\\s]+").ReplaceAllString(s, "") //spaces
	s = regexp.MustCompile("\\s+").ReplaceAllString(s, "-")         //spaces
	return s
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

//GenerateRandomStringURLSafe generates a secure random string of atleast n bytes (actually more because of base64)
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}
