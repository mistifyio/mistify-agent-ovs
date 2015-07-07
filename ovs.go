package ovs

func ovs(args ...string) ([][]string, error) {
	cmd := command{command: "ovs-vsctl"}
	return cmd.Run(append([]string{"--format", "json"}, args...)...)
}

func listOVSIfaces(bridge string) ([]*Iface, error) {
	args := []string{
		"list-ifaces",
		bridge,
	}

	resultLines, err := ovs(args...)
	if err != nil {
		return nil, err
	}

	ifaces := make([]*Iface, len(resultLines))
	for i, resultLine := range resultLines {

	}
}
func getOVSIface(bridge, ifaceName string) (*Iface, error) {
}
func addOVSIface(bridge, ifaceName string) (*Iface, error) {
}
func removeOVSIface(bridge, ifaceName string) error {
}
