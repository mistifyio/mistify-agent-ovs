/*
Package ovs is a mistify subagent that manages virtual network interfaces. It
uses systemd-networkd to create new interfaces and joins them to a bridge using
Open vSwitch.

HTTP API Endpoints

	/_mistify_RPC_
		* GET - Run a specified method

Request Structure

	{
		"method": "OVS.RPC_METHOD",
		"params": [
			DATA_STRUCT
		],
		"id": 0
	}

Where RPC_METHOD is the desired method and DATA_STRUCTURE is an rpc.GuestRequest
http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestRequest

Response Structure

	{
		"result": {
			"guest": RESPONSE_STRUCT
		},
		"error":null,
		"id":0
	}

Where RESPONSE_STRUCT is an rpc.GuestResponse containing changes in the guest
nics array.
http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestResponse

RPC Methods

	AddGuestInterfaces
	RemoveGuestInterfaces

See the godocs for each method's purpose
*/
package ovs
