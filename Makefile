PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin

cmd/mistify-operator-admin/mistify-operator-admin: cmd/mistify-operator-admin/main.go
	cd cmd/mistify-operator-admin && \
	go get && \
	go build

clean:
	cd cmd/mistify-operator-admin && \
	go clean

install: cmd/mistify-operator-admin/mistify-operator-admin
	install -D cmd/mistify-operator-admin/mistify-operator-admin $(DESTDIR)$(SBIN_DIR)/mistify-operator-admin
