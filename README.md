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

    AddGuestInterface
    RemoveGuestInterface

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
func NewOVS(bridge string) *OVS
```
NewOVS creates a new OVS object

#### func (*OVS) AddGuestInterface

```go
func (ovs *OVS) AddGuestInterface(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
AddGuestInterface creates a new interface for a guest

#### func (*OVS) RemoveGuestInterface

```go
func (ovs *OVS) RemoveGuestInterface(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error
```
RemoveGuestInterface removes the interface for a guest

#### func (*OVS) RunHTTP

```go
func (ovs *OVS) RunHTTP(port uint) error
```
RunHTTP creates and runs the RPC HTTP server

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
