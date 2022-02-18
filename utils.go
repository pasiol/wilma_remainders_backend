package main

import (
	"context"
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/pasiol/wilma_remainders_backend/configs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang.org/x/crypto/bcrypt"
)

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SplitOrigins() ([]string, error) {
	s, exists := os.LookupEnv("APP_ALLOWED_ORIGINS")
	if !exists {
		return []string{}, errors.New("ALLOWED_ORIGINS variable missing")
	}
	origins := strings.Split(s, ",")

	for _, origin := range origins {
		uri, err := url.ParseRequestURI(origin)
		if err == nil && (uri.Scheme != "https" && uri.Scheme != "http") {
			return []string{}, errors.New("malformed uri")
		}
	}
	return origins, nil
}

func GetDebug() bool {
	return os.Getenv("APP_DEBUG") == "true"
}

func connectOrFail(uri string, db string) (*mongo.Database, *mongo.Client, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}
	err = client.Connect(context.Background())
	if err != nil {
		return nil, nil, err
	}
	var DB = client.Database(db)
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, nil, err
	}

	return DB, client, nil
}

func transformRemainder(r Remainder) ([]Remainder, error) {
	roleRegexp, err := regexp.Compile(`([a-z]+)`)
	idRegexp, err := regexp.Compile(`([0-9]+)`)
	if err != nil {
		return []Remainder{}, nil
	}
	trasformed := []Remainder{}
	splitted := strings.Split(r.To, "#role#")
	for i, recipient := range splitted {
		if recipient != "" {
			recipientData := strings.SplitN(splitted[i], "@", 2)

			if len(recipientData) == 2 {
				role := string(roleRegexp.Find([]byte(recipientData[0])))
				id := string(idRegexp.Find([]byte(recipientData[0])))
				if slugRole, found := configs.Roles[role]; found {
					email := recipientData[1]
					slug := slugRole + id
					r := Remainder{
						To:      email,
						Title:   r.Title,
						Message: strings.Replace(r.Message, "#SLUG#", slug, -1),
						// TODO: types
						UpdatedAt: r.UpdatedAt,
					}
					trasformed = append(trasformed, r)
					r = Remainder{}
				}
			} else {
				return []Remainder{}, errors.New("malformed recipient data")
			}
		}
	}
	return trasformed, nil
}

func clean(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}

func sanitizeSearch(s string) (string, error) {
	searchRegexp, err := regexp.Compile(`[0-9A-Za-z.@-]+`)
	if err != nil {
		return "", err
	}
	sanitized := searchRegexp.Find([]byte(s))
	return string(sanitized), nil
}
