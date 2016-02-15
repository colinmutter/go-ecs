package main

import (
	"errors"
	"fmt"
	"os/user"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-ini/ini"
)

var profileName string

// GetSessionWithProfile will return a new aws session with optional profile
func GetSessionWithProfile(profile string) (*session.Session, error) {
	config := aws.NewConfig()

	// Get and save profile
	if profile != "" {
		profileName = profile
		config = config.WithCredentials(credentials.NewSharedCredentials("", profileName))

		// Debug creds...
		// fmt.Println(config.Credentials.Get())
	}

	return getSessionWithConfig(config)
}

// GetSession starts an AWS session with region auto-detection
func GetSession() (*session.Session, error) {
	config := aws.NewConfig()
	return getSessionWithConfig(config)
}

// getSessionWithConfig grabs the region and appends to the current config
func getSessionWithConfig(config *aws.Config) (*session.Session, error) {
	region, err := getAWSRegion()

	if profileName != "" {
		fmt.Println("Using profile: ", *profile)
	} else {
		fmt.Println("Using profile: default")
	}

	if region != "" {
		fmt.Println("With region: ", region)
		config = config.WithRegion(region)
	}
	return session.New(config), err
}

func getAWSRegion() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	cfgPath := fmt.Sprintf("%s/.aws/config", currentUser.HomeDir)
	cfg, err := ini.Load(cfgPath)

	if err != nil {
		return "", err
	}

	var sectionName string
	if profileName != "" {
		sectionName = fmt.Sprintf("profile %s", profileName)
	} else {
		sectionName = "default"
	}

	section := cfg.Section(sectionName)
	if err != nil || !section.HasKey("region") {
		return "", errors.New("Did not find AWS region from config file")
	}

	region := section.Key("region").String()

	return region, nil
}
