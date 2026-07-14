# Installation and Configuration Guide
MAP can be installed by one of three approaches: 

- **Online installer:** The installer downloads MAP's images from our official repository. For this reason, the installer is very small in size.

- **Offline installer:** Use this installer when the host does not have an Internet connection. The installer contains pre-built images so its size is larger.

- **Virtual Appliance:** If you are using MAP as a instance on vSphere platform, download the OVA version of MAP. 

All installers can be downloaded from the **[official release](https://gitlab.eng.vmware.com/cn-se-solution/mapd/-/tags)** page. 

To install MAP's virtual appliance, refer to the **[Installation Guide for Virtual Appliance](installation_guide_ova.md)**.

This guide describes the steps to install and configure MAP by using the online or offline installer. The installation processes are almost the same. 

If you run a previous version of MAP, you may need to migrate the data to fit the new database schema. For more details, please refer to **[Upgrade Guide](upgrade_guide.md)**.

## Prerequisites for the target host
MAP is deployed as several Docker containers, and, therefore, can be deployed on any Linux distribution that supports Docker. The target host requires Docker, and Docker Compose to be installed.
* Docker engine should be version 1.10 or higher.  For installation instructions, please refer to: https://docs.docker.com/engine/installation/
* Docker Compose needs to be version 1.6.0 or higher.  For installation instructions, please refer to: https://docs.docker.com/compose/install/

## Installation Steps

The installation steps boil down to the following

1. Download the installer;
2. Configure **config.yml**;
3. Run **install.sh** to install and start MAP;


### Downloading the installer:

The binary of the installer can be downloaded from the [release](https://gitlab.eng.vmware.com/cn-se-solution/mapd/-/tags) page. Choose either online or offline installer. Use *tar* command to extract the package.

Online installer:
```
    $ tar -xvf map-online-installer-<version>.tgz
```
Offline installer:
```
    $ tar -xvf map-offline-installer-<version>.tgz
```

### Configuring MAP
Configuration parameters are located in the file **config.yml**. 
The parameters are described below - note that at the very least, you will need to change the **hostname** attribute. 

* **hostname**: The target host's hostname, which is used to access the UI and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, e.g., `192.168.1.10` or `map.yourdomain.com`. _Do NOT use `localhost`, `127.0.0.1`, or other illegal FQDN like `map` for the hostname - the endpoint needs to be accessible by external browser!_ 
* **http.port**: The HTTP service port. Requests sent to this port will be automatically redirected to the HTTPS port.
* **https.port**: The HTTPS service port. Default is 443.
* **https.certificate**: The TLS certificate file.
* **https.private_key**: The TLS private key file.
* **https.ca_certificate**: The root CA certificate file. This is required when `https.certificate` is issued by untrusted 3rd party certificate authority. Also, MAP will use this CA certificate to verify all HTTPS connections. The CA certificate file can contains multiple PEM blocks in it.

* **map_admin_password**: The default administrator's password. This password takes effect everytime MAP launches. _Note that the default administrator's username is **admin@tanzu.local**._   
* **enable_cross_client_token_sharing**: If it is allowed to authorize with token that is issued to other IDP clients. The server could be vulnerable when this is enabled. This is a client-side mechanism that will not affect IDP server's behaviour.
* **timezone**: Timezone of the running containers

* **database.password**: The root password for the MySQL database used for mapd. _Change this password for any production use!_ 
* **database.max_idle_conns**: The maximum number of connections in the idle connection pool. Do not change this unless you met performance issues.
* **database.max_open_conns**: The maximum number of open connections to the database. Do not change this unless you met performance issues.
* **data_volume**: The default data volume is `/data`. Do not use symbolic link on the volume folder.
* **log.level**: The logging level.
* **log.rotate_count**: Log files are rotated log_rotate_count times before being removed. If count is 0, old versions are removed rather than rotated.
* **log.rotate_size**: Log files are rotated only if they grow bigger than log_rotate_size bytes. If size is followed by k, the size is assumed to be in kilobytes.
  If the M is used, the size is in megabytes, and if G is used, the size is in gigabytes. So size 100, size 100k, size 100M and size 100G are all valid.
* **log.location**: The directory on your host that store log.

### Finishing installation and starting MAP
Once **config.yml** are configured, install and start MAP using the ```install.sh``` script. Note that it may take some time for the online installer to download images from Official registry.  

```sh
$ sudo ./install.sh
```

If everything worked properly, you should be able to open a browser to visit the admin web console at **http://map.yourdomain.com** (change *map.yourdomain.com* to the hostname configured in your config.yml). Note that the default administrator username are `admin@tanzu.local`.

Log in to the admin web console and add a image repository, e.g. `myharbor.yourdomain.com`. To use this image repository for initializing kubernetes cluster, you will need to upload MAP dependency images to this image repository.

```sh
$ sudo ./download_dependency.sh
$ sudo docker login myharbor.yourdomain.com
$ sudo ./upload_dependency.sh -r myharbor.yourdomain.com
```

For information on how to use Harbor, please refer to **[User Guide of Harbor](user_guide.md)** .

### Configuring authorization with external IDP

MAP uses [Dex](https://dexidp.io/) for its identity management. Dex is an open source [OpenID Connect](https://openid.net/connect/) identity provider. It can be configured manually to authenticate against external IDPs with [connectors](https://dexidp.io/docs/connectors/). By default, Dex is configured to use `local` connector only.

To add additional connectors to dex, please follow the below steps:

1. Stop service with command: `docker-compose down -v`.
2. Edit the dex configuration template file: `vi common/config/dex/config-template.yaml`.

	Keep in mind that the following block **SHALL NOT** be changed:
	```
	issuer: ${DEX_ISSUER}
	storage:
	  type: mysql
	  config:
	    host: ${MYSQL_HOST}
	    port: ${MYSQL_PORT}
	    database: ${MYSQL_DEX_DATABASE}
	    user: ${MYSQL_USER}
	    password: ${MYSQL_PASSWORD}
	    ssl:
	      mode: "false"
	web:
	  http: 0.0.0.0:5556
	telemetry:
	  http: 0.0.0.0:5558
	grpc:
	  addr: 0.0.0.0:5557
	frontend:
	  issuer: ${MAP_ISSUER}
	  dir: /opt/dex/web-${DEX_PAGE_LANG}
	  theme: map
	oauth2:
	  skipApprovalScreen: true
	enablePasswordDB: ${DEX_LOCAL_CONNECTOR_ENABLED}
	expiry:
	  deviceRequests: ${DEX_EXPIRY_DEVICE_REQUESTS}
	  signingKeys: ${DEX_EXPIRY_SIGNING_KEYS}
	  idTokens: ${DEX_EXPIRY_ID_TOKENS}
	  authRequests: ${DEX_EXPIRY_AUTH_REQUESTS}
	logger:
	  level: ${DEX_LOG_LEVEL}
	  format: ${DEX_LOG_FORMAT}
	```

	Add an additional connector like this:
	```
	issuer: ${DEX_ISSUER}
	storage:
	  type: mysql
	  config:
	    host: ${MYSQL_HOST}
	    port: ${MYSQL_PORT}
	    database: ${MYSQL_DEX_DATABASE}
	    user: ${MYSQL_USER}
	    password: ${MYSQL_PASSWORD}
	    ssl:
	      mode: "false"
	web:
	  http: 0.0.0.0:5556
	telemetry:
	  http: 0.0.0.0:5558
	grpc:
	  addr: 0.0.0.0:5557
	frontend:
	  issuer: ${MAP_ISSUER}
	  dir: /opt/dex/web-${DEX_PAGE_LANG}
	  theme: map
	oauth2:
	  skipApprovalScreen: true
	enablePasswordDB: ${DEX_LOCAL_CONNECTOR_ENABLED}
	expiry:
	  deviceRequests: ${DEX_EXPIRY_DEVICE_REQUESTS}
	  signingKeys: ${DEX_EXPIRY_SIGNING_KEYS}
	  idTokens: ${DEX_EXPIRY_ID_TOKENS}
	  authRequests: ${DEX_EXPIRY_AUTH_REQUESTS}
	logger:
	  level: ${DEX_LOG_LEVEL}
	  format: ${DEX_LOG_FORMAT}
	staticClients:
	connectors:
	# Example of LDAP connector. Full documentation can be found here: https://dexidp.io/docs/connectors.
	- type: ldap
	  id: ldap
	  name: Microsoft AD
	  config:
	    host: 10.1.2.20:636
	    # The following field is required if using port 389.
	    #insecureNoSSL: true
	    #rootCA: tmp/rootca.pem
	    insecureSkipVerify: true
	    bindDN: CN=Service Account,OU=dev,DC=example,DC=org
	    bindPW: P@ssw0rd
	    usernamePrompt: Email Address
	    userSearch:
	      baseDN: OU=dev,DC=example,DC=org
	      filter: "(objectClass=person)"
	      username: userPrincipalName
	      idAttr: DN
	      emailAttr: userPrincipalName
	      nameAttr: name
	      preferredUsernameAttr: sAMAccountName
	    groupSearch:
	      baseDN: OU=dev,DC=example,DC=org
	      filter: "(|(objectClass=posixGroup)(objectClass=group))"
	      userMatchers:
	      - userAttr: DN
	        groupAttr: member
	      nameAttr: cn
	```
3. Start service with command: `docker-compose up -d`

	When MAP does not work properly, run the below commands to find out if all containers of MAP are in **UP** status: 
	```sh
	$ docker-compose ps
	NAME                COMMAND                  SERVICE             STATUS              PORTS
	map-console         "/docker-entrypoint.…"   map-console         running (healthy)   0.0.0.0:8080->8080/tcp, 0.0.0.0:8443->8443/tcp, :::8080->8080/tcp, :::8443->8443/tcp
	map-dex             "/bin/bash -c /opt/d…"   dex                 running (healthy)   127.0.0.1:5556-5558->5556-5558/tcp
	map-log             "/bin/sh -c /usr/loc…"   log                 running (healthy)   127.0.0.1:4514->10514/tcp
	map-mapd            "/usr/bin/entrypoint…"   mapd                running (healthy)   127.0.0.1:9800->9800/tcp
	map-mariadb         "docker-entrypoint.s…"   mariadb             running (healthy)   127.0.0.1:3306->3306/tcp                         
	```

	If a container is not in **UP** state, check the log file of that container. For example, if the container ```map-dex``` is not running, you can check the logging message with the following command for troubleshooting:
	
	```sh
	docker logs map-dex
	```

## Managing MAP's lifecycle
You can use docker-compose to manage the lifecycle of MAP. Some useful commands are listed as follows (must run in the same directory as *docker-compose.yml*).

Stopping MAP:
```sh
$ sudo docker-compose stop
Stopping map-console ... done
Stopping map-mapd ... done
Stopping map-dex ... done
Stopping map-mariadb ... done
Stopping map-log ... done
```
Restarting MAP after stopping:
```sh
$ sudo docker-compose start
Starting map-log ... done
Starting map-mariadb ... done
Starting map-dex ... done
Starting map-mapd ... done
Starting map-console ... done
```  

To change MAP's configuration, first stop existing MAP instance, update config.yml, and then run install.sh again:
```sh
$ sudo docker-compose down -v
$ vim config.yml
$ sudo install.sh
``` 

Removing MAP's containers while keeping MAP's database files on the file system:
```
$ sudo docker-compose rm
Going to remove map-console, map-mapd, map-dex, map-mariadb
Are you sure? [yN] y
Removing map-console ... done
Removing map-mapd ... done
Removing map-dex ... done
Removing map-mariadb ... done
Removing map-log ... done
```  

Removing MAP's database data (for a clean re-installation, the path may be different depends on the value of `data_volume` in `config.yml`):
```sh
$ rm -rf /data/mariadb
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

### Persistent data and log files
By default, registry data is persisted in the target host's `/data/` directory.  This data remains unchanged even when MAP's containers are removed and/or recreated.  

In addition, MAP uses rsyslog to collect the logs of each container. By default, these log files are stored in the directory /var/log/map/ on the target host for troubleshooting.

## Troubleshooting

1. When MAP does not work properly, run the below commands to find out if all containers of MAP are in **UP** status: 
```sh
$ sudo docker-compose ps
NAME                COMMAND                  SERVICE             STATUS              PORTS
map-console         "/docker-entrypoint.…"   map-console         running (healthy)   0.0.0.0:8080->8080/tcp, 0.0.0.0:8443->8443/tcp, :::8080->8080/tcp, :::8443->8443/tcp
map-dex             "/bin/bash -c /opt/d…"   dex                 running (healthy)   127.0.0.1:5556-5558->5556-5558/tcp
map-log             "/bin/sh -c /usr/loc…"   log                 running (healthy)   127.0.0.1:4514->10514/tcp
map-mapd            "/usr/bin/entrypoint…"   mapd                running (healthy)   127.0.0.1:9800->9800/tcp
map-mariadb         "docker-entrypoint.s…"   mariadb             running (healthy)   127.0.0.1:3306->3306/tcp                         
```

If a container is not in **UP** state, check the log file of that container. For example, if the container ```map-dex``` is not running, you can check the logging message with the following command:

```sh
sudo docker logs map-dex
```

2. When setting up MAP behind an nginx proxy or elastic load balancing, please keep the frontend hostname unchanged between MAP and external LB.

A wrong configuration:
```
                                  External LB                        MAP
                                       │                              │
  request ─────────────1──────────────>│──────────────2──────────────>│
                                       │                              │
      Hostname: host1.yourdomain.com   │   Hostname: map.example.org  │
                                       │                              │
                                       │<─────────────3───────────────│response
                                       │                              │
                                       │       Error: HTTP 422        │
```

A correct configuration:
```
                                  External LB                        MAP
                                       │                              │
  request ─────────────1──────────────>│──────────────2──────────────>│
                                       │                              │
           Hostname: map.example.org   │   Hostname: map.example.org  │
                                       │                              │
                                       │<─────────────3───────────────│response
                                       │                              │
                                       │       OK: HTTP 200           │
```