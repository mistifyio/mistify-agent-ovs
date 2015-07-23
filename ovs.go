package ovs

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	// ovsIfaceData is a struct to help in decoding interface related output
	// from ovs-vsctl
	ovsIfaceData struct {
		Headings []string
		Data     [][]string
	}

	// OVS is the Mistify OS subagent service
	OVS struct {
		bridge          string
		ifacePrefix     string
		lastIfaceNumber int
		maxIfaceNumber  int
	}
)

// ovsIfaceListJSONtoNic converts the json output of an
// `ovs-vsctl list Interface` command to an array of client.Nic objects
func ovsIfaceListJSONToNic(bridge, input string) ([]*client.Nic, error) {
	ifaceData := &ovsIfaceData{}
	if err := json.Unmarshal([]byte(input), ifaceData); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"input": input,
		}).Error("failed to parse ovs iface json")
		return nil, err
	}

	// Get a mapping of header name to array index
	headings := make(map[string]int)
	for i, colname := range ifaceData.Headings {
		headings[colname] = i
	}

	// Assemble results
	results := make([]*client.Nic, len(ifaceData.Data))
	for i, row := range ifaceData.Data {
		results[i] = &client.Nic{
			Name:    row[headings["name"]],
			Device:  row[headings["name"]],
			Network: bridge,
			Mac:     row[headings["mac_in_use"]],
		}
	}

	return results, nil
}

// cleanNic removes interface details from a nic object, leaving the externally
// configured details like bridge and IP
func cleanNic(nic client.Nic) client.Nic {
	nic.Name = ""
	nic.Device = ""
	nic.Mac = ""
	return nic
}

// ovs is a convenience wrapper for running ovs-vsctl commands
func ovs(args ...string) ([]string, error) {
	cmd := command{command: "ovs-vsctl"}
	return cmd.Run(append([]string{"--format", "json"}, args...)...)
}

// listOVSIfaces gets a list of interfaces attached to a bridge
func listOVSIfaces(bridge string) ([]*client.Nic, error) {
	args := []string{
		"list-ifaces",
		bridge,
	}

	ifaceNames, err := ovs(args...)
	if err != nil {
		return nil, err
	}

	if len(ifaceNames) == 0 {
		return []*client.Nic{}, nil
	}

	return getOVSIfaces(bridge, ifaceNames...)
}

// getOVSIfaces gets one or more interfaces attached to a bridge, by name
func getOVSIfaces(bridge string, ifaceNames ...string) ([]*client.Nic, error) {
	args := []string{
		"--columns",
		"name,mac_in_use,type",
		"list",
		"Interface",
	}
	args = append(args, ifaceNames...)
	results, err := ovs(args...)
	if err != nil {
		return nil, err
	}

	// The json formatting for list returns the entire result on one line
	ifaceListJSON := results[0]
	return ovsIfaceListJSONToNic(bridge, ifaceListJSON)
}

// addOVSIface attaches an interface to a bridge
func addOVSIface(bridge, ifaceName string, vlanTag int) error {
	if vlanTag <= 0 {
		vlanTag = 1
	}
	args := []string{
		"add-port",
		bridge,
		ifaceName,
		fmt.Sprintf("tag=%d", vlanTag),
	}
	_, err := ovs(args...)
	return err
}

// deleteOVSIface removes an interface from a bridge
func deleteOVSIface(bridge, ifaceName string) error {
	args := []string{
		"del-port",
		bridge,
		ifaceName,
	}
	_, err := ovs(args...)
	return err
}

// NewOVS creates a new OVS object
func NewOVS(bridge string) (*OVS, error) {
	ovs := &OVS{
		bridge:      bridge,
		ifacePrefix: "mist",
	}
	// 15 characters, minus the prefix and separator
	// e.g. No prefix, "." sep = 10^(15-1) - 1 = 99999999999999 = 14 characters
	ovs.maxIfaceNumber = int(math.Pow10(15-(len(ovs.ifacePrefix)+1)) - 1)

	// Find how high the numbering has gone
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		parts := strings.SplitN(iface.Name, ".", 2)
		if len(parts) != 2 || parts[0] != ovs.ifacePrefix {
			continue
		}
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		if i > ovs.lastIfaceNumber {
			ovs.lastIfaceNumber = i
		}
	}

	return ovs, nil
}

// newIfaceName generates a new interface name
func (ovs *OVS) newIfaceName() (string, error) {
	initial := ovs.lastIfaceNumber
	newIfaceNumber := ovs.lastIfaceNumber + 1
	ifaceName := ""
	for newIfaceNumber != initial {
		// Too high, start back at 0
		if newIfaceNumber > ovs.maxIfaceNumber {
			newIfaceNumber = 0
		}

		ifaceName = fmt.Sprintf("%s.%d", ovs.ifacePrefix, newIfaceNumber)

		// Check whether an interface with that name already exists
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil && err.Error() != "no such network interface" {
			log.WithFields(log.Fields{
				"error":     err,
				"ifaceName": ifaceName,
			}).Error("failed to look up interface by name")
			return "", err
		}
		ovs.lastIfaceNumber = newIfaceNumber
		if iface == nil {
			return ifaceName, nil
		}
		newIfaceNumber++
	}
	err := errors.New("no free interface names")
	log.WithFields(log.Fields{
		"error":          err,
		"ifacePrefix":    ovs.ifacePrefix,
		"maxIfaceNumber": ovs.maxIfaceNumber,
	}).Error("unable to generate interface name")
	return "", err
}

// AddGuestInterfaces creates new interfaces for a guest
func (ovs *OVS) AddGuestInterfaces(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	if request.Guest == nil || request.Guest.Nics == nil {
		return errors.New("missing guest with nics")
	}
	for i, nic := range request.Guest.Nics {
		if nic.Name != "" {
			continue
		}

		ifaceName, err := ovs.newIfaceName()
		if err != nil {
			return err
		}

		bridge := nic.Network
		if bridge == "" {
			bridge = ovs.bridge
		}

		// Create TAP interface
		if err := createTAPIface(ifaceName); err != nil {
			return err
		}

		// Add TAP interface to OVS
		if err := addOVSIface(bridge, ifaceName, 0); err != nil {
			// Clean up
			_ = deleteTAPIface(ifaceName)
			return err
		}

		// Populate the guest nic with new info
		iface, err := net.InterfaceByName(ifaceName)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"ifaceName": ifaceName,
			}).Error("failed to look up interface by name")
			return err
		}
		request.Guest.Nics[i].Name = iface.Name
		request.Guest.Nics[i].Device = iface.Name
		request.Guest.Nics[i].Network = bridge
		request.Guest.Nics[i].Mac = iface.HardwareAddr.String()
	}
	response.Guest = request.Guest
	return nil
}

// RemoveGuestInterfaces removes the interfaces for a guest
func (ovs *OVS) RemoveGuestInterfaces(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	if request.Guest == nil || request.Guest.Nics == nil {
		return errors.New("missing guest with nics")
	}
	for i, nic := range request.Guest.Nics {
		if nic.Name == "" {
			continue
		}

		// Remove TAP interface from OVS
		if err := deleteOVSIface(nic.Network, nic.Name); err != nil {
			return err
		}

		// Remove TAP Interface
		if err := deleteTAPIface(nic.Name); err != nil {
			return err
		}

		request.Guest.Nics[i] = cleanNic(nic)
	}

	response.Guest = request.Guest
	return nil
}

// requestIfaceName determines the interface name from guest request information
func requestIfaceName(request *rpc.GuestRequest) (string, error) {
	if request.Guest == nil || request.Guest.Id == "" {
		return "", errors.New("missing guest with id")
	}
	return strings.Split(request.Guest.Id, "-")[4], nil
}
