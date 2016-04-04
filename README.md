# A Highly Available Kubernetes PostgreSQL Setup

###### This is experimental

#### K8s:
* Set up the Kubernetes controller manager with `--pod-eviction-timeout=30s` or something reasonable, as this will be the trigger for failover during a network outage where the Master is running.
* Create a `postgresqlmaster` service before creating the `ReplicationController` or `DaemonSet`. This will ensure the proper environment variables will be created in the containers below.
* An optional read-only `postgresqlslave` service can be created to read the slave instances in round robin.
* Master and Slave share the same volume (probably `hostPath`.)
* Master/Slaves can be started in either order.

### Environment Variable settings:
| Env Var | Example Values |
| --- | --- |
| POSTGRES_MODE | master or slave |
| POSTGRES_BASE_DIR | /data/postgres |
| SLAVE_TRIGGER_FILE | postgresql.trigger |
| SHUTDOWN_SLAVE_FILE | SHUTDOWN_SLAVE |
| SHUTDOWN_SLAVE_SUCCESS_FILE | SHUTDOWN_SLAVE_SUCCESS |
| POSTGRES_ENTRYPOINT | /usr/lib/postgresql/9.5/bin/postgres |
| POSTGRES_OPTIONS | "-D $PGDATA -c config_file=/data/postgres/conf/postgresql.conf"


### Currently we shell-out and rely on the following commands:
* `rm`
* `touch`

## Master `ReplicationController`:
#### K8s:
* Set `replicas: 1`
* Use `PreStop` to `rm /data/postgres/SHUTDOWN_SLAVE /data/postgres/SHUTDOWN_SLAVE_SUCCESSFUL` (in case a node comes back online after a network outage)

Logic:
* Ensure `POSTGRES_MODE` is set
* `touch /data/postgres/data/postgresql.trigger`? ( TODO: check if this is needed...)
* `touch /data/postgres/SHUTDOWN_SLAVE`
* Wait for `/data/postgres/SHUTDOWN_SLAVE_SUCCESSFUL` to appear
* `rm /data/postgres/data/postgresql.trigger -rf`? (TODO: check if this is needed...)
* Start postgresql


## Slave `DaemonSet`:
#### K8s:
* Use healthCheck for reinitiation

#### Logic:
* If `/data/postgres/SHUTDOWN_SLAVE` exists:
> * Watch for `/data/postgres/SHUTDOWN_SLAVE` and `/data/postgres/SHUTDOWN_SLAVE_SUCCESSFUL` to disappear
> * Wait for positive health check from postgres master (reached via service, using env vars `POSTGRESQLMASTER_SERVICE_HOST` and `POSTGRESQLMASTER_SERVICE_PORT`)
> * Write `/data/postgres/data/recovery.conf`
> * Run `pg_basebackup`
> * Start postgresql
* Else:
> * Wait for positive health check from postgres master (reached via service, using env vars `POSTGRESQLMASTER_SERVICE_HOST` and `POSTGRESQLMASTER_SERVICE_PORT`)
> * Write `/data/postgres/data/recovery.conf`
> * Run `pg_basebackup`
> * Start postgresql

* Watch for `/data/postgres/SHUTDOWN_SLAVE`, if it appears:
> * Shutdown postgresql and wait for full exit
> * `touch /data/postgres/SHUTDOWN_SLAVE_SUCCESSFUL`

## TODO
* convert to tmpfs for testing file actions
* convert to native go instead of shelling-out
