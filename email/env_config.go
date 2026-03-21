package email

import (
	"fmt"
	"html/template"
	"time"

	"github.com/matcornic/hermes/v2"
)

type (
	// EmailBranding holds app-specific branding used when building Hermes email templates.
	EmailBranding struct {
		CompanyName string
		LogoURL     string
	}

	// EnvironmentConfig is the configuration for a given environment.
	EnvironmentConfig struct {
		baseURL template.URL
		outboundInvitesEmailAddress,
		passwordResetCreationEmailAddress,
		passwordResetRedemptionEmailAddress string
	}
)

// BaseURL returns the BaseURL field.
func (c *EnvironmentConfig) BaseURL() template.URL {
	return c.baseURL
}

// OutboundInvitesEmailAddress returns the OutboundInvitesEmailAddress field.
func (c *EnvironmentConfig) OutboundInvitesEmailAddress() string {
	return c.outboundInvitesEmailAddress
}

// PasswordResetCreationEmailAddress returns the passwordResetCreationEmailAddress field.
func (c *EnvironmentConfig) PasswordResetCreationEmailAddress() string {
	return c.passwordResetCreationEmailAddress
}

// PasswordResetRedemptionEmailAddress returns the passwordResetRedemptionEmailAddress field.
func (c *EnvironmentConfig) PasswordResetRedemptionEmailAddress() string {
	return c.passwordResetRedemptionEmailAddress
}

func (c *EnvironmentConfig) BuildHermes(branding *EmailBranding) *hermes.Hermes {
	var name, logo, copyright string
	if branding != nil {
		name = branding.CompanyName
		logo = branding.LogoURL
		copyright = fmt.Sprintf("Copyright © %d %s. All rights reserved.", time.Now().Year(), branding.CompanyName)
	}
	return &hermes.Hermes{
		Product: hermes.Product{
			Name:      name,
			Link:      string(c.baseURL),
			Logo:      logo,
			Copyright: copyright,
		},
	}
}
