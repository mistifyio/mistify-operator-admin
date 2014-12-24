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

create_test_db_user:
	sudo -u postgres psql -c "create user testoperator with superuser password 'testpass';"

delete_test_db_user:
	-sudo -u postgres dropuser testoperator

create_test_db:
	sudo -u postgres createdb testdb && \
	sudo -u postgres psql -q testdb < schema.sql

delete_test_db:
	-sudo -u postgres dropdb testdb

test_setup: create_test_db_user create_test_db

test_clean: delete_test_db_user delete_test_db

test_config: cmd/mistify-operator-admin/testconfig.json
	cd config; \
	go test

test_db: cmd/mistify-operator-admin/testconfig.json
	cd db; \
	go test

test : | test_setup test_config test_db test_clean
