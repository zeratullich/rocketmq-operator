## Deploy NFS Service 

This document will show you how to simply deploy the NFS shared storage service on your Kubernetes cluster with an example.

### Install NFS Server

Choose one of your nodes to install the NFS service, assuming our shared data directory is: `/data/nfs/k8s`

1. Turn off the firewall (requires sudo privileges)

```
$ systemctl stop firewalld.service

$ systemctl disable firewalld.service
```

2. Install and configure NFS

```
$ yum -y install nfs-utils rpcbind
```

3. Set permissions for the shared directory:

```
$ chmod 755 /data/nfs/k8s/
```

4. Configure NFS. The default configuration file for NFS is under `/etc/exports`. Add the following configuration information to the file:

```
$ vi /etc/exports
/data/nfs/k8s   192.168.0.0/16(rw,sync,no_root_squash,no_all_squash)
```

5. Start rpcbind service:

```
$ systemctl start rpcbind.service

$ systemctl enable rpcbind

$ systemctl status rpcbind
···
Jun 25 18:11:50 docker-nfs systemd[1]: Starting RPC bind service...
Jun 25 18:11:51 docker-nfs systemd[1]: Started RPC bind service.
```

If you see the above output, the rpcbind is successfully started.

6. Start NFS service:

```
$ systemctl start nfs.service

$ systemctl enable nfs

$ systemctl status nfs
···
Jun 25 18:11:56 docker-nfs systemd[1]: Starting NFS server and services...
Jun 25 18:11:57 docker-nfs systemd[1]: Started NFS server and services.
```

If you see the above output, the NFS server is successfully started.

It can also be confirmed by the following command:

```
$ rpcinfo -p | grep nfs
    100003    3   tcp   2049  nfs
    100003    4   tcp   2049  nfs
    100227    3   tcp   2049  nfs_acl
    100003    3   udp   2049  nfs
    100003    4   udp   2049  nfs
    100227    3   udp   2049  nfs_acl
```

So here we have the NFS server installed, then we install the NFS service on the other nodes.

### Install NFS Clients

Install NFS on the other nodes of your Kubernetes cluster:

1. Turn off the firewall first:

```
$ systemctl stop firewalld.service

$ systemctl disable firewalld.service
```

2. Install NFS 

```
$ yum -y install nfs-utils rpcbind
```

3. As with the above method, start rpc first, then start NFS:

```
$ systemctl start rpcbind.service 

$ systemctl enable rpcbind.service 

$ systemctl start nfs.service    

$ systemctl enable nfs.service
```

4. After the client startup is complete, check if NFS has a shared directory:

```
$ showmount -e 192.168.0.250
Export list for 192.168.0.250:
/data/nfs/k8s 192.168.0.0/16
```

Where `192.168.0.250` is the node IP where your NFS server is located.

For further verification operations, please see [NFS Storage Preparation Document](../README.md)