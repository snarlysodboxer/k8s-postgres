# A Highly Available Kubernetes PostgreSQL Setup

###### This is experimental

#### K8s:
* Setup the Kubernetes controller manager with `--pod-eviction-timeout=30s` or something reasonable, as this will be the trigger for failover during a network outage where the Master is running.
* Create a `postgresql_master` `Service` before creating the `ReplicationController` or `DaemonSet`. This will ensure the proper environment variables will be created in the containers below.
* An optional read-only `postgresqlslave` service can be created to read the slave instances in round robin.
* Master and Slave share the same data volume (probably `hostPath`, could be a data container) where postgres data is stored.
* Master/Slaves can be started in either order.
* Master:

> * Use `ReplicationController`
> * Set `replicas: 1`
* Slave:

> * Use `DaemonSet`
> * Use `healthCheck` for reinitiation
> * Set pod IP in `/data/postgres/slave_ip` using:
```
  volumeMounts:
    - name: podip
      mountPath: /data/postgres
      readOnly: true
  volumes:
    - name: podip
      downwardAPI:
        items:
          - path: "slave_ip"
            fieldRef:
              fieldPath: status.podIP
```

### Environment Variables:
| Env Var | Example Value |
| --- | --- |
| POSTGRES_MODE | master or slave |
| POSTGRES_BASE_DIR | /data/postgres |
| POSTGRES_TRIGGER_FILE | postgresql.trigger |
| POSTGRES_ENTRYPOINT | /usr/lib/postgresql/9.5/bin/postgres |
| POSTGRES_OPTIONS | "-D $PGDATA -c config_file=/data/postgres/conf/postgresql.conf" |

### Currently we shell-out and rely on the following commands:
* `touch`

#### Logic:
* Ensure `POSTGRES_MODE` is set
* If master

> * Create trigger file
> * Check for and wait for `/data/postgres/slave_ip` to show
> * Port Forward from `DaemonSet` ("slave") through `ReplicationController` ("master") - get IP from `cat /data/postgres/slave_ip`
* If slave

> * Wait for positive health check from postgres master (reached via service, using env vars `POSTGRESQL_MASTER_SERVICE_HOST` and `POSTGRESQL_MASTER_SERVICE_PORT`)
> * Create `recovery.conf` file
> * Run `pg_basebackup`
> * Start PostgreSQL

## TODO
* convert to tmpfs for testing file actions
* convert to native go instead of shelling-out
