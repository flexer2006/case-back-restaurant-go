package notification_test

import (
	"errors"
	"testing"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	"github.com/flexer2006/case-back-restaurant-go/internal/notification"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

type MockDialer struct {
	mock.Mock
}

func (m *MockDialer) DialAndSend(messages ...*gomail.Message) error {
	args := m.Called(messages)
	return args.Error(0)
}

type MockNetError struct {
	timeout   bool
	temporary bool
	msg       string
}

func (e *MockNetError) Error() string   { return e.msg }
func (e *MockNetError) Timeout() bool   { return e.timeout }
func (e *MockNetError) Temporary() bool { return e.temporary }

type MockOpError struct {
	Op  string
	Net string
	Err error
}

func (e *MockOpError) Error() string { return e.Op + " " + e.Net + ": " + e.Err.Error() }

func createTestSMTPConfig() *configs.SMTPConfig {
	return &configs.SMTPConfig{
		Host:     "smtp.yandex.ru",
		Port:     465,
		Username: "test@yandex.ru",
		Password: "testpass",
		From:     "test@yandex.ru",
		Secure:   true,
	}
}

func TestNewSMTPMailer(t *testing.T) {
	t.Run("with valid config", func(t *testing.T) {
		config := createTestSMTPConfig()
		mailer := notification.NewSMTPMailer(config)
		assert.NotNil(t, mailer, "SMTPMailer should not be nil")
	})

	t.Run("with nil config", func(t *testing.T) {
		mailer := notification.NewSMTPMailer(nil)
		assert.NotNil(t, mailer, "SMTPMailer should not be nil even with nil config")

		err := mailer.SendEmail("test@example.com", "Test", "Test body")
		assert.Error(t, err, "Sending with nil config should return error")
		assert.Contains(t, err.Error(), common.ErrInitSMTP)
	})
}

func TestSendEmail(t *testing.T) {

	originalDialerFactory := notification.NewDialer
	defer func() {
		notification.NewDialer = originalDialerFactory
	}()

	t.Run("successful email sending", func(t *testing.T) {
		config := createTestSMTPConfig()
		mockDialer := new(MockDialer)

		notification.NewDialer = func(host string, port int, username, password string) notification.Dialer {
			return mockDialer
		}

		mockDialer.On("DialAndSend", mock.Anything).Return(nil)

		mailer := notification.NewSMTPMailer(config)
		err := mailer.SendEmail("recipient@example.com", "Test Subject", "Test Body")

		assert.NoError(t, err)
		mockDialer.AssertExpectations(t)
	})

	t.Run("validation errors", func(t *testing.T) {
		config := createTestSMTPConfig()
		mailer := notification.NewSMTPMailer(config)

		err := mailer.SendEmail("", "Subject", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), common.ErrEmptyRecipient)

		err = mailer.SendEmail("test@example.com", "", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), common.ErrEmptySubject)

		err = mailer.SendEmail("invalid-email", "Subject", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), common.ErrSMTPInvalidRecipient)
	})

	t.Run("network timeout error", func(t *testing.T) {
		config := createTestSMTPConfig()
		mockDialer := new(MockDialer)

		notification.NewDialer = func(host string, port int, username, password string) notification.Dialer {
			return mockDialer
		}

		netErr := &MockNetError{timeout: true, msg: "connection timeout"}
		mockDialer.On("DialAndSend", mock.Anything).Return(netErr)

		mailer := notification.NewSMTPMailer(config)
		err := mailer.SendEmail("recipient@example.com", "Test Subject", "Test Body")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), common.ErrDialSMTP)
		assert.Contains(t, err.Error(), common.ErrSMTPTimeout)
		mockDialer.AssertExpectations(t)
	})

	t.Run("network operation error", func(t *testing.T) {
		config := createTestSMTPConfig()
		mockDialer := new(MockDialer)

		notification.NewDialer = func(host string, port int, username, password string) notification.Dialer {
			return mockDialer
		}

		simpleErr := errors.New("dial tcp: connection refused")
		mockDialer.On("DialAndSend", mock.Anything).Return(simpleErr)

		mailer := notification.NewSMTPMailer(config)
		err := mailer.SendEmail("recipient@example.com", "Test Subject", "Test Body")

		assert.Error(t, err)

		assert.Contains(t, err.Error(), common.ErrSendEmail)
		mockDialer.AssertExpectations(t)
	})

	t.Run("context timeout", func(t *testing.T) {
		config := createTestSMTPConfig()
		mockDialer := new(MockDialer)

		notification.NewDialer = func(host string, port int, username, password string) notification.Dialer {
			return mockDialer
		}

		mockDialer.On("DialAndSend", mock.Anything).Run(func(args mock.Arguments) {

			time.Sleep(100 * time.Millisecond)
		}).Return(nil)

		originalSendTimeout := notification.SMTPSendTimeout
		notification.SMTPSendTimeout = 50 * time.Millisecond
		defer func() {
			notification.SMTPSendTimeout = originalSendTimeout
		}()

		mailer := notification.NewSMTPMailer(config)
		err := mailer.SendEmail("recipient@example.com", "Test Subject", "Test Body")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), common.ErrSendEmail)
		assert.Contains(t, err.Error(), common.ErrSMTPTimeout)
		mockDialer.AssertExpectations(t)
	})
}
