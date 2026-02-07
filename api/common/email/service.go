package email

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"sync-backend/api/common/email/model"
	"sync-backend/arch/config"
	"sync-backend/arch/mongo"
	"sync-backend/arch/network"
	"sync-backend/utils"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailService interface {
	SendPasswordReset(email, resetToken, resetUrl string) error
	SendEmailVerification(email, verificationToken, verificationUrl string) error
	SendWelcomeEmail(email, username string) error
}

type emailService struct {
	network.BaseService
	logger         utils.AppLogger
	env            *config.Env
	queryBuilder   mongo.QueryBuilder[model.EmailLog]
	sendgridClient *sendgrid.Client
	fromEmail      string
	fromName       string
}

func NewEmailService(env *config.Env, db mongo.Database) EmailService {
	client := sendgrid.NewSendClient(env.SendGridAPIKey)

	return &emailService{
		BaseService:    network.NewBaseService(),
		logger:         utils.NewServiceLogger("EmailService"),
		env:            env,
		queryBuilder:   mongo.NewQueryBuilder[model.EmailLog](db, model.EmailLogCollectionName),
		sendgridClient: client,
		fromEmail:      env.SendGridFromEmail,
		fromName:       env.SendGridFromName,
	}
}

func (s *emailService) SendPasswordReset(email, resetToken, resetUrl string) error {
	subject := "Reset Your Password - Sync"

	// Render HTML template
	htmlContent, err := s.renderTemplate("password_reset.html", map[string]interface{}{
		"ResetUrl": resetUrl,
		"Email":    email,
	})
	if err != nil {
		s.logger.Error("Failed to render password reset template: %v", err)
		return err
	}

	// Send email
	err = s.sendEmail(email, subject, htmlContent, model.EmailTypePasswordReset)
	if err != nil {
		s.logger.Error("Failed to send password reset email to %s: %v", email, err)
		return err
	}

	s.logger.Success("Password reset email sent successfully to: %s", email)
	return nil
}

func (s *emailService) SendEmailVerification(email, verificationToken, verificationUrl string) error {
	subject := "Verify Your Email - Sync"

	// Render HTML template
	htmlContent, err := s.renderTemplate("email_verification.html", map[string]interface{}{
		"VerificationUrl": verificationUrl,
		"Email":           email,
	})
	if err != nil {
		s.logger.Error("Failed to render email verification template: %v", err)
		return err
	}

	// Send email
	err = s.sendEmail(email, subject, htmlContent, model.EmailTypeVerification)
	if err != nil {
		s.logger.Error("Failed to send email verification to %s: %v", email, err)
		return err
	}

	s.logger.Success("Email verification sent successfully to: %s", email)
	return nil
}

func (s *emailService) SendWelcomeEmail(email, username string) error {
	subject := "Welcome to Sync!"

	// Render HTML template
	htmlContent, err := s.renderTemplate("welcome.html", map[string]interface{}{
		"Username": username,
		"Email":    email,
	})
	if err != nil {
		s.logger.Error("Failed to render welcome email template: %v", err)
		return err
	}

	// Send email
	err = s.sendEmail(email, subject, htmlContent, model.EmailTypeWelcome)
	if err != nil {
		s.logger.Error("Failed to send welcome email to %s: %v", email, err)
		return err
	}

	s.logger.Success("Welcome email sent successfully to: %s", email)
	return nil
}

func (s *emailService) sendEmail(to, subject, htmlContent string, emailType model.EmailType) error {
	// Create email log
	emailLog := model.NewEmailLog(to, subject, emailType)

	// Create SendGrid message
	from := mail.NewEmail(s.fromName, s.fromEmail)
	recipient := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, recipient, "", htmlContent)

	// Send email via SendGrid
	response, err := s.sendgridClient.Send(message)
	if err != nil {
		// Log failure
		emailLog.Status = model.EmailStatusFailed
		emailLog.Error = err.Error()
		s.logEmail(emailLog)
		return err
	}

	// Check response status
	if response.StatusCode >= 400 {
		errMsg := fmt.Sprintf("SendGrid returned status %d: %s", response.StatusCode, response.Body)
		emailLog.Status = model.EmailStatusFailed
		emailLog.Error = errMsg
		s.logEmail(emailLog)
		return fmt.Errorf("%s", errMsg)
	}

	// Log success
	now := primitive.NewDateTimeFromTime(time.Now())
	emailLog.Status = model.EmailStatusSent
	emailLog.SentAt = &now
	s.logEmail(emailLog)

	return nil
}

func (s *emailService) renderTemplate(templateName string, data map[string]interface{}) (string, error) {
	// Get the absolute path to templates directory
	templatesDir := filepath.Join("api", "common", "email", "templates")
	templatePath := filepath.Join(templatesDir, templateName)

	// Parse and execute template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Execute template to string
	var result string
	builder := &stringBuilder{}
	err = tmpl.Execute(builder, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	result = builder.String()
	return result, nil
}

func (s *emailService) logEmail(emailLog *model.EmailLog) {
	_, err := s.queryBuilder.Query(context.Background()).InsertOne(emailLog)
	if err != nil {
		s.logger.Error("Failed to log email: %v", err)
	}
}

// stringBuilder is a simple wrapper to implement io.Writer for template execution
type stringBuilder struct {
	content string
}

func (sb *stringBuilder) Write(p []byte) (n int, err error) {
	sb.content += string(p)
	return len(p), nil
}

func (sb *stringBuilder) String() string {
	return sb.content
}
