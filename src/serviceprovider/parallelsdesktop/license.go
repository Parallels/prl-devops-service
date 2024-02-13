package parallelsdesktop

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func (s *ParallelsService) GetLicense() (*models.ParallelsDesktopLicense, error) {
	getLicenseCmd := helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"info", "--license", "--json"},
	}

	output, _, _, err := helpers.ExecuteWithOutput(getLicenseCmd)
	if err != nil {
		return nil, err
	}

	output = strings.ReplaceAll(output, "This feature is not available in this edition of Parallels Desktop. \n", "")

	var license models.ParallelsDesktopLicense
	err = json.Unmarshal([]byte(output), &license)
	if err != nil {
		return nil, err
	}

	return &license, nil
}

func (s *ParallelsService) InstallLicense(licenseKey, username, password string) error {
	if licenseKey == "" {
		return errors.New("license key is required")
	}

	installLicenseCmd := helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"install-license", "-k", licenseKey, "--activate-online-immediately"},
	}

	if username != "" && password != "" {
		passwordCmd := helpers.Command{
			Command: "echo",
			Args:    []string{password, ">~/parallels_password.txt"},
		}
		signInCmd := helpers.Command{
			Command: s.serverExecutable,
			Args:    []string{"web-portal", "signin", username, "--read-passwd", "~/parallels_password.txt"},
		}
		_, err := helpers.ExecuteWithNoOutput(passwordCmd)
		if err != nil {
			return err
		}
		_, err = helpers.ExecuteWithNoOutput(signInCmd)
		if err != nil {
			return err
		}
		_, err = helpers.ExecuteWithNoOutput(installLicenseCmd)
		if err != nil {
			return err
		}

		return nil
	} else {
		_, err := helpers.ExecuteWithNoOutput(installLicenseCmd)
		if err != nil {
			return err
		}

		return nil
	}
}

func (s *ParallelsService) DeactivateLicense() error {
	logger.Info("Deactivating Parallels Desktop license")
	deactivateLicenseCmd := helpers.Command{
		Command: s.serverExecutable,
		Args:    []string{"deactivate-license", "--skip-network-errors"},
	}

	_, err := helpers.ExecuteWithNoOutput(deactivateLicenseCmd)
	if err != nil {
		return err
	}

	logger.Info("Parallels Desktop license deactivated successfully")

	return nil
}
