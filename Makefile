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

create_db_user:
	-sudo -u postgres createuser -l operator;

test_setup: create_db_user
	sudo -u postgres createdb testdb && \
	sudo -u postgres psql -q testdb < schema.sql

test_clean:
	-sudo -u postgres dropdb testdb

test_config: cmd/mistify-operator-admin/testconfig.json
	cd config; \
	go test

test : | test_setup test_config test_clean
