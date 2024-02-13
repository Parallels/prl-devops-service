package password

import "github.com/Parallels/pd-api-service/basecontext"

type PasswordComplexityOptions struct {
	ctx                      basecontext.ApiContext
	minLength                int
	maxLength                int
	requireLowercase         bool
	requireUppercase         bool
	requireNumbers           bool
	requireSpecialCharacters bool
	saltPassword             bool
}

func NewPasswordComplexityOptions(ctx basecontext.ApiContext) *PasswordComplexityOptions {
	return &PasswordComplexityOptions{
		ctx:                      ctx,
		minLength:                12,
		maxLength:                40,
		requireLowercase:         true,
		requireUppercase:         true,
		requireNumbers:           true,
		requireSpecialCharacters: true,
		saltPassword:             true,
	}
}

func (p *PasswordComplexityOptions) WithMinLength(minLength int) *PasswordComplexityOptions {
	if minLength < 8 {
		p.ctx.LogWarnf("Password complexity options MinLength cannot be less than 8. Setting to 8.")
		minLength = 8
	}
	p.minLength = minLength
	return p
}

func (p *PasswordComplexityOptions) WithMaxLength(maxLength int) *PasswordComplexityOptions {
	if maxLength > 40 {
		p.ctx.LogWarnf("Password complexity options MaxLength cannot be greater than 40. Setting to 40.")
		maxLength = 40
	}

	p.maxLength = maxLength
	return p
}

func (p *PasswordComplexityOptions) WithRequireLowercase(requireLowercase bool) *PasswordComplexityOptions {
	p.requireLowercase = requireLowercase
	return p
}

func (p *PasswordComplexityOptions) WithRequireUppercase(requireUppercase bool) *PasswordComplexityOptions {
	p.requireUppercase = requireUppercase
	return p
}

func (p *PasswordComplexityOptions) WithRequireNumbers(requireNumbers bool) *PasswordComplexityOptions {
	p.requireNumbers = requireNumbers
	return p
}

func (p *PasswordComplexityOptions) WithRequireSpecialCharacters(requireSpecialCharacters bool) *PasswordComplexityOptions {
	p.requireSpecialCharacters = requireSpecialCharacters
	return p
}

func (p *PasswordComplexityOptions) WithSaltPassword(saltPassword bool) *PasswordComplexityOptions {
	p.saltPassword = saltPassword
	return p
}

func (p *PasswordComplexityOptions) MinLength() int {
	return p.minLength
}

func (p *PasswordComplexityOptions) MaxLength() int {
	return p.maxLength
}

func (p *PasswordComplexityOptions) RequireLowercase() bool {
	return p.requireLowercase
}

func (p *PasswordComplexityOptions) RequireUppercase() bool {
	return p.requireUppercase
}

func (p *PasswordComplexityOptions) RequireNumbers() bool {
	return p.requireNumbers
}

func (p *PasswordComplexityOptions) RequireSpecialCharacters() bool {
	return p.requireSpecialCharacters
}

func (p *PasswordComplexityOptions) SaltPassword() bool {
	return p.saltPassword
}
