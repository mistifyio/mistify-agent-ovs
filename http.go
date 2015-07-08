package ovs

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
)

// RunHTTP creates and runs the RPC HTTP server
func (ovs *OVS) RunHTTP(port uint) error {
	server, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to create rpc server")
		return err
	}
	if err := server.RegisterService(ovs); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to register OVS service")
		return err
	}

	if err := server.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to run server")
		return err
	}
	return nil
}
