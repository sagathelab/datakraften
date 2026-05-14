package docker

import (
	"os"
	"strings"

	"github.com/sagathelab/datakraften/internal/exec"
)

type DockerStatus struct {
	CliInstalled bool
	DaemonRunning bool
	DockerSocket bool
	WSLIntegration bool
	Message      string
}

func Detect() DockerStatus {
	status := DockerStatus{}

	cliPath := exec.CommandPath("docker")
	status.CliInstalled = cliPath != ""

	_, err := os.Stat("/var/run/docker.sock")
	status.DockerSocket = err == nil

	if status.CliInstalled {
		r := exec.Run("docker", "info")
		status.DaemonRunning = r.Code == 0

		if status.DaemonRunning && strings.Contains(r.Stdout, "WSL2") {
			status.WSLIntegration = true
		}
	}

	if !status.CliInstalled {
		status.Message = "Docker CLI not found"
	} else if !status.DaemonRunning {
		status.Message = "Docker daemon not running — enable Docker Desktop WSL integration"
	} else if status.WSLIntegration {
		status.Message = "Docker Desktop WSL integration active"
	} else {
		status.Message = "Docker available (integration status unknown)"
	}

	return status
}

func HasDockerSocket() bool {
	_, err := os.Stat("/var/run/docker.sock")
	return err == nil
}
