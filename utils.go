package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

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

	trasformed := []Remainder{}
	//fmt.Printf("\n%v\n", r.To)
	splitted := strings.Split(r.To, "#role#")
	//fmt.Printf("\n%v\n", splitted)
	for i, recipient := range splitted {
		if recipient != "" {
			fmt.Printf("\n%s\n", splitted[i])
			recipientData := strings.SplitN(splitted[i], "@", 2)

			if len(recipientData) == 2 {
				role := recipientData[0]
				slug := ""
				if strings.Contains(role, "student") || strings.Contains(role, "parent") {
					slug = strings.Replace(role, "student", "", -1)
					slug = strings.Replace(slug, "parent", "", -1)
					slug = slug + "/"
				}
				email := recipientData[1]
				r := Remainder{
					To:      email,
					Title:   r.Title,
					Message: strings.Replace(r.Message, "#SLUG#/", slug, -1),
					// TODO: types
					UpdatedAt: r.UpdatedAt,
				}
				trasformed = append(trasformed, r)
				r = Remainder{}
			} else {
				return []Remainder{}, errors.New("malformed recipient data")
			}
		}
	}
	//splitted = nil
	return trasformed, nil
}
