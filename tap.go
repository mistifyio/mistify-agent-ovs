package ovs

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

// tapTemplate is the NetDev file template for a TAP interface
var tapTemplate = `
[NetDev]
Name=%s
Kind=tap

[Tap]
`

// netDevDir is the directory systemd-networkd checks for new interface definitions
var netDevDir = "/etc/systemd/network"

// netDevFilepath builds the full path for a NetDev file
func netDevFilepath(ifaceName string) string {
	return filepath.Join(netDevDir, fmt.Sprintf("%s.netdev", ifaceName))
}

// createTAPIface creates a new TAP interface
func createTAPIface(ifaceName string) error {
	filepath := netDevFilepath(ifaceName)

	// Create the netdev file
	file, err := os.Create(filepath)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filepath": filepath,
		}).Error("failed to create netdev file")
		return err
	}
	defer file.Close()

	// Populate the file from the template. If there's a problem, clean up
	contents := fmt.Sprintf(tapTemplate, ifaceName)
	_, err = file.Write([]byte(contents))
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filepath": filepath,
			"contents": contents,
		}).Error("failed to write netdev file")
		_ = os.Remove(filepath)
		return err
	}

	// Restart networkd so changes take effect
	return restartNetworkd()
}

// deleteTAPIface removes a TAP interface
func deleteTAPIface(ifaceName string) error {
	filepath := netDevFilepath(ifaceName)

	// Remove netdev file
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		log.WithFields(log.Fields{
			"error":    err,
			"filepath": filepath,
		}).Error("failed to remove netdev file")
		return err
	}

	// Remove the interface
	cmd := command{command: "ip"}
	args := []string{
		"link",
		"delete",
		ifaceName,
	}
	if _, err := cmd.Run(args...); err != nil {
		return err
	}

	// Restart networkd so changes take effect
	return restartNetworkd()
}

// restartNetworkd restarts systemd-networkd
func restartNetworkd() error {
	cmd := command{command: "systemctl"}
	_, err := cmd.Run("restart", "systemd-networkd")
	if err != nil {
		return err
	}
	return nil
}
