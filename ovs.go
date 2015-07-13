package ovs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		bridge string
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
func addOVSIface(bridge, ifaceName string, vlanTag int) (*client.Nic, error) {
	if vlanTag <= 0 {
		vlanTag = 1
	}
	args := []string{
		"add-port",
		bridge,
		ifaceName,
		fmt.Sprintf("tag=%d", vlanTag),
	}
	if _, err := ovs(args...); err != nil {
		return nil, err
	}

	nics, err := getOVSIfaces(bridge, ifaceName)
	if err != nil {
		return nil, err
	}
	return nics[0], nil
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
func NewOVS(bridge string) *OVS {
	return &OVS{
		bridge: bridge,
	}
}

// AddGuestInterface creates a new interface for a guest
func (ovs *OVS) AddGuestInterface(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	ifaceName, err := requestIfaceName(request)
	if err != nil {
		return err
	}

	// Create TAP interface
	if err := createTAPIface(ifaceName); err != nil {
		return err
	}

	// Add TAP interface to OVS
	nic, err := addOVSIface(ovs.bridge, ifaceName, 0)
	if err != nil {
		// Clean up
		_ = deleteTAPIface(ifaceName)
		return err
	}

	response.Guest = request.Guest
	response.Guest.Nics = []client.Nic{*nic}
	return nil
}

// RemoveGuestInterface removes the interface for a guest
func (ovs *OVS) RemoveGuestInterface(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	ifaceName, err := requestIfaceName(request)
	if err != nil {
		return err
	}

	// Remove TAP interface from OVS
	if err := deleteOVSIface(ovs.bridge, ifaceName); err != nil {
		return err
	}

	// Remove TAP Interface
	if err := deleteTAPIface(ifaceName); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.Nics = []client.Nic{}
	return nil
}

// requestIfaceName determines the interface name from guest request information
func requestIfaceName(request *rpc.GuestRequest) (string, error) {
	if request.Guest == nil || request.Guest.Id == "" {
		return "", errors.New("missing guest with id")
	}
	return strings.Split(request.Guest.Id, "-")[4], nil
}
