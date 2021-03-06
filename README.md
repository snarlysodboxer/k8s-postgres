# A Highly Available Kubernetes PostgreSQL Setup

###### This is experimental, a proof of concept

## The Goal
- Have a single PostgreSQL master running anywhere in the K8s cluster
- Always reach the master through the same IP; use a K8s `Service`
- Have one Streaming Replication slave on each labeled `minion`, providing redundancy and distributed read-only; use K8s `DaemonSet`
- Failover re-connects all slaves to the new master
- Let K8s `ReplicationController` choose where the master goes
- Be provider agnostic (Don't use GCE Persistent Volumes or AWS EBS Volumes)
- Don't reinvent the wheel, use K8s as much as possible

## How we do it

### Kubernetes Setup
- Start the Controller Manager with `--pod-eviction-timeout=30s` or something reasonable, as this will be the trigger for failover during a network level outage where the master is running.
- Create a `postgres_master` `Service` before creating the `ReplicationController` or `DaemonSet` to ensure the proper environment variables will be created in the containers.
- An optional read-only `postgres_slave` service can be created to read the slave instances in round robin.
- Master and Slave share the same data volume (probably `hostPath`, could be a data container) where postgres data is stored.
- Master/Slaves can be started in either order.

- Master:
  - Use `ReplicationController`
  - Set `replicas: 1`
  - Set `containerPort: 5432`

- Slave:
  - Use `DaemonSet` (1 per `minion`)
  - Set `healthCheck` for reinitiation (checking port will catch crashed/stopped postgres, but will miss out of sync.)
  - Set `containerPort: 5433`
  - Set pod IP in `/data/postgres/slave_ip` using the `DownwardAPI`

### Docker Image Setup
- Build the `pg-controller` binary, add it to a PostgreSQL image, set it as the Entrypoint
- Build in tcp-proxy at `/tcp-proxy` - https://github.com/lumanetworks/go-tcp-proxy
- Currently we shell-out and rely on the Unix commands `touch` and `nc`
- Environment Variables

  | Variable | Example Value |
| --- | --- |
| POSTGRES_MODE | master or slave |
| POSTGRES_TRIGGER_FILE | /data/postgres/data/postgresql.trigger |
| POSTGRES_ENTRYPOINT | /usr/lib/postgresql/9.5/bin/postgres |
| POSTGRES_OPTIONS | "-D /data/postgres/data -c config_file=/data/postgres/conf/postgresql.conf" |
| POSTGRES_SLAVE_IP_FILE | /data/postgres/slave_ip |
| POSTGRES_DATA_DIR | /data/postgres/data |
| POSTGRES_RECOVERY_FILE | /data/postgres/data/recovery.conf |
| POSTGRES_MASTER_SERVICE_HOST | set automatically by the "postgres_master" service |
| POSTGRES_MASTER_SERVICE_PORT | set automatically by the "postgres_master" service |
| POSTGRES_REPLICATOR_USER | replicator |
| POSTGRES_REPLICATOR_PASS | your pass |

### Runtime Logic:
- Ensure `POSTGRES_MODE` is set to determine master vs slave
- If master
  - Create trigger file
  - Check for and wait for `/data/postgres/slave_ip` to show
  - Reverse Proxy from `DaemonSet` ("slave") through `ReplicationController` ("master") - get IP from `/data/postgres/slave_ip`

- If slave
  - Wait for either
    - Positive health check from postgres master (reached via service, using env vars `POSTGRES_MASTER_SERVICE_HOST` and `POSTGRES_MASTER_SERVICE_PORT`), *OR*
    - Trigger file to exist or appear
  - When Positive health check
    - Create `recovery.conf` file
    - Run `pg_basebackup`
    - Start PostgreSQL as slave
  - When Trigger file
    - Start PostgreSQL as master

## TODO
- convert to tmpfs for testing file actions
- convert to native go instead of shelling-out
- create proxy instead of using someone else's
- consider using the k8s api to change the `Endpoint` of a static `Service` - instead of proxying directly
- create a template file that can be customized instead of using Sprintf
- consider settings other than environment variables - make it better
- consider running pg_basebackup only when needed instead of every time
