package services

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetSlurmToken 为指定用户生成token
func GetSlurmToken(username, lifespanSec string) (string, error) {
	cmd := exec.Command("scontrol", "token", fmt.Sprintf("username=%s", username), fmt.Sprintf("lifespan=%s", lifespanSec))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("scontrol command failed: %w, output: %s", err, string(output))
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "SLURM_JWT=") {
		return strings.TrimPrefix(outputStr, "SLURM_JWT="), nil
	}

	return "", fmt.Errorf("unexpected output from scontrol: %s", outputStr)
}
