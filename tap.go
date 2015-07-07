package ovs

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

var tapTemplate = `
[NetDev]
Name=%s
Kind=tap

[Tap]
`

var netDevPath = "/etc/systemd/network"

func netDevFilepath(ifaceName string) string {
	return filepath.Join(netDevPath, fmt.Sprintf("%s.netdev", ifaceName))
}

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

	// Populate the file from the template. If there's a problem, clean up
	contents := fmt.Sprintf(tapTemplate, ifaceName)
	_, err := file.Write([]byte(contents))
	file.Close()
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

	// Restart networkd so changes take effect
	return restartNetworkd()
}

func restartNetworkd() error {
	cmd := command{"command": "systemctl"}
	output, err := cmd.Run("restart", "systemd-networkd")
	if err != nil {
		return err
	}
	return nil
}
