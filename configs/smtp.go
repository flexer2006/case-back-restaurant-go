package configs

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/flexer2006/case-back-restaurant-go/common"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Secure   bool
}

func NewSMTPConfig() (*SMTPConfig, error) {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	secureStr := os.Getenv("SMTP_SECURE")

	var missingParams []string
	if host == "" {
		missingParams = append(missingParams, "SMTP_HOST")
	}
	if portStr == "" {
		missingParams = append(missingParams, "SMTP_PORT")
	}
	if username == "" {
		missingParams = append(missingParams, "SMTP_USERNAME")
	}
	if password == "" {
		missingParams = append(missingParams, "SMTP_PASSWORD")
	}
	if from == "" {
		missingParams = append(missingParams, "SMTP_FROM")
	}

	if len(missingParams) > 0 {
		return nil, fmt.Errorf(common.ErrSMTPRequiredParams, strings.Join(missingParams, ", "))
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrSMTPPortParse, err)
	}

	if port <= 0 {
		return nil, errors.New(common.ErrSMTPInvalidPort)
	}

	secure := true
	if secureStr != "" {
		secure, err = strconv.ParseBool(secureStr)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", common.ErrSMTPSecureParse, err)
		}
	}

	if !strings.Contains(from, "@") {
		return nil, fmt.Errorf("%s: %s", common.ErrSMTPInvalidConfig, common.ErrSMTPInvalidSenderEmail)
	}

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		Secure:   secure,
	}, nil
}
