set up the Kubernetes controller manager with `--pod-eviction-timeout=30s` or something more reasonable for failover during node failure

create a 'postgresqlmaster' service before starting the ReplicationController or DaemonSet (so the proper environment variables will be created in the containers below)
an optional 'postgresqlslave' service can be created to read the read-only slave instances in round robin
master and slave share the same volume (probably hostPath)
master/slaves can be started in either order

  master ReplicationController:
replicas: 1
use PreStop to `rm /postgres/SHUTDOWN_SLAVE /postgres/SHUTDOWN_SLAVE_SUCCESSFUL` (in case a node comes back online after a network outage)

ensure POSTGRES_MODE is set
`touch /postgres/data/postgresql.trigger`? (not sure if this is needed...)
`touch /postgres/SHUTDOWN_SLAVE`
wait for `/postgres/SHUTDOWN_SLAVE_SUCCESSFUL` to appear
`rm /postgres/data/postgresql.trigger -rf`? (not sure if this is needed...)
start postgresql as master


  slave DaemonSet:
uses slave healthCheck for reinitiation

if `/postgres/SHUTDOWN_SLAVE` exists:
  watch for `/postgres/SHUTDOWN_SLAVE` and `/postgres/SHUTDOWN_SLAVE_SUCCESSFUL` to disappear
  wait for positive health check from postgres master (reached via service, using env vars `POSTGRESQLMASTER_SERVICE_HOST` and `POSTGRESQLMASTER_SERVICE_PORT`)
  write `/postgres/data/recovery.conf`
  run `pg_basebackup`
  start postgresql as slave
else:
  wait for positive health check from postgres master (reached via service, using env vars `POSTGRESQLMASTER_SERVICE_HOST` and `POSTGRESQLMASTER_SERVICE_PORT`)
  write `/postgres/data/recovery.conf`
  run `pg_basebackup`
  start postgresql as slave
watch for `/postgres/SHUTDOWN_SLAVE`, if it appears:
  shutdown postgresql and wait for full exit
  `touch /postgres/SHUTDOWN_SLAVE_SUCCESSFUL`

#### TODO
* convert to tmpfs for testing file actions
* use native go instead of shelling-out

### Currently we shell-out and rely on the following commands:
* rm
* touch
