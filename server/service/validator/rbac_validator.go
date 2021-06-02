package validator

import "github.com/go-chassis/cari/rbac"

func ValidateCreateAccount(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return createAccountValidator.Validate(a)
}

func ValidateUpdateAccount(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return updateAccountValidator.Validate(a)
}
func ValidateCreateRole(a *rbac.Role) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return createRoleValidator.Validate(a)
}
func ValidateAccountLogin(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return accountLoginValidator.Validate(a)
}
func ValidateChangePWD(a *rbac.Account) error {
	err := baseCheck(a)
	if err != nil {
		return err
	}
	return changePWDValidator.Validate(a)
}
