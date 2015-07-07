PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin
ETC_DIR=$(PREFIX)/etc

cmd/mistify-agent-ovs/mistify-agent-ovs: cmd/mistify-agent-ovs/main.go
	cd cmd/mistify-agent-ovs && \
	go get -v && \
	go build -v

clean:
	cd cmd/mistify-agent-ovs && \
	go clean -v

install: cmd/mistify-agent-ovs/mistify-agent-ovs
	install -D cmd/mistify-agent-ovs/mistify-agent-ovs $(DESTDIR)$(SBIN_DIR)/mistify-agent-ovs
