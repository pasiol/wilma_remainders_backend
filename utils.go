package main

import (
	"context"
	"errors"
	"log"
	"net/url"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"../configs"

	"golang.org/x/crypto/bcrypt"
)

func hashAndSalt(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

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

func getRecipients(s string) []Recipient {
	var recipients []Recipient
	splittedRecipients := strings.Split(s, "#role#")

	for i, r := range splittedRecipients[1:] {
		fields := strings.SplitN(r, "@", 2)
		recipient := Recipient{}
		if len(fields) == 2 {
			if strings.Contains(r, "teacher") {
				recipient = Recipient{
					Email:          fields[1],
					Role:           "teacher",
					SlugIdentifier: strings.Replace(fields[0], "teacher", configs.TeacherRole, 1),
				}
				if configs.TeacherRole == "" {
					recipient.SlugIdentifier = ""
				}
			}
			if strings.Contains(r, "student") {
				recipient = Recipient{
					Email:          fields[1],
					Role:           "student",
					SlugIdentifier: strings.Replace(fields[0], "student", configs.StudentRole, 1),
				}
				if configs.StudentRole == "" {
					recipient.SlugIdentifier = ""
				}
			}
			if strings.Contains(r, "personel") {
				recipient = Recipient{
					Email:          fields[1],
					Role:           "personel",
					SlugIdentifier: strings.Replace(fields[0], "personel", configs.PersonelRole, 1),
				}
				if configs.PersonelRole == "" {
					recipient.SlugIdentifier = ""
				}
			}
			if strings.Contains(r, "parent") {
				recipient = Recipient{
					Email:          fields[1],
					Role:           "parent",
					SlugIdentifier: strings.Replace(fields[0], "parent", configs.ParentRole, 1),
				}
				if configs.ParentRole == "" {
					recipient.SlugIdentifier = ""
				}
			}
			if strings.Contains(r, "nowilma-account") {
				recipient = Recipient{
					Email:          fields[1],
					Role:           "parent",
					SlugIdentifier: strings.Replace(fields[0], "nowilma-account", configs.Anonymous, 1),
				}
				if configs.ParentRole == "" {
					recipient.SlugIdentifier = ""
				}
			}
			recipients = append(recipients, recipient)
		} else {
			log.Printf("Splittig recipient failed: %d, %s: %v", i, r, fields)
		}
	}
	return recipients
}
