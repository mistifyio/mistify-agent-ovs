/*
mistify-agent-ovs runs the subagent and HTTP API.

Usage

The following arguments are understood:

	$ mistify-agent-ovs -h
	Usage of mistify-agent-ovs:
	-b, --bridge="mistify0": bridge to join interfaces to with OVS
	-l, --log-level="warning": log level: debug/info/warning/error/fatal
	-p, --port=40001: listen port
*/
package main
