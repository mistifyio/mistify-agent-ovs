# ovs

[![ovs](https://godoc.org/github.com/mistifyio/mistify-agent-ovs?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent-ovs)

Package ovs is a mistify subagent that manages virtual network interfaces. It
uses systemd-networkd to create new interfaces and joins them to a bridge using
Open vSwitch.

### HTTP API Endpoints

    /_mistify_RPC_
    	* GET - Run a specified method

### Request Structure

    {
    	"method": "OVS.RPC_METHOD",
    	"params": [
    		DATA_STRUCT
    	],
    	"id": 0
    }

Where RPC_METHOD is the desired method and DATA_STRUCTURE is an rpc.GuestRequest
http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestRequest

### Response Structure

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

### RPC Methods

    AddGuestInterfaces
    RemoveGuestInterfaces

See the godocs for each method's purpose

## Usage

#### type OVS

```go
type OVS struct {
}
```

OVS is the Mistify OS subagent service

#### func  NewOVS

```go
func NewOVS(bridge string) (*OVS, error)
```
NewOVS creates a new OVS object

#### func (*OVS) AddGuestInterfaces

```go
func (ovs *OVS) AddGuestInterfaces(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
AddGuestInterfaces creates new interfaces for a guest

#### func (*OVS) RemoveGuestInterfaces

```go
func (ovs *OVS) RemoveGuestInterfaces(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
RemoveGuestInterfaces removes the interfaces for a guest

#### func (*OVS) RunHTTP

```go
func (ovs *OVS) RunHTTP(port uint) error
```
RunHTTP creates and runs the RPC HTTP server

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
