# RocketMQ Operator
> Note: 
+ Forked by [Apache RocketMQ Operator](https://github.com/apache/rocketmq-operator) .
+ Change 60% of the source code and fix code bugs(
Fix the problem that the broker only can be increased and cannot be reduced sizes.) .
+ Change use Service instead of node address .
+ Can only be used in k8s cluster .
## Quick Start
### Deploy RocketMQ Operator
1. Clone the project on your Kubernetes cluster master node:
``` 
$ git clone https://github.com/zeratullich/rocketmq-operator
$ cd rocketmq-operator 
```
2. To deploy the RocketMQ Operator on your Kubernetes cluster, please run the following script:
```
$ kubectl create ns rocketmq
$ kubectl ns rocketmq
$ ./install-operator.sh
```
3. Use command `kubectl get pods` to check the RocketMQ Operator deploy status like:
```
$ kubectl get pods 
NAME                                 READY   STATUS    RESTARTS   AGE
rocketmq-operator-6cb8f7d6c4-79m2j   1/1     Running   0          102s
```
### Prepare Volume Persistence
Before RocketMQ deployment, you may need to do some preparation steps for RocketMQ data persistence.

Currently we provide several options for your RocketMQ data persistence: `EmptyDir`, `HostPath` and `NFS`, which can be configured in CR files, for example in `rocketmq.zeratullich.org_v1beta1_nameservice_cr.yaml`:
```
...
 # storageMode can be EmptyDir, HostPath, NFS
  storageMode: NFS
...
```
If you choose `EmptyDir`, you don't need to do extra preparation steps for data persistence. But the data storage life is the same with the pod's life, if the pod is deleted you may lost the data.

If you choose other storage modes, please refer to the following instructions to prepare the data persistence.
#### Prepare HostPath
This storage mode means the RocketMQ data (including all the logs and store files) is stored in each host where the pod lies on. In that case you need to create an dir where you want the RocketMQ data to be stored on.

We provide a script in `deploy/storage/hostpath/prepare-host-path.sh`, which you can use to create the `HostPath` dir on every worker node of your Kubernetes cluster.
```
$ cd deploy/storage/hostpath
$ sudo su
$ ./prepare-hostpath.sh 
```
Changed hostPath /data/rocketmq/nameserver uid to 3000, gid to 3000
Changed hostPath /data/rocketmq/broker uid to 3000, gid to 3000
You may refer to the instructions in the script for more information.
#### Prepare Storage Class of NFS
If you choose NFS as the storage mode, the first step is to prepare a storage class based on NFS provider to create PV and PVC where the RocketMQ data will be stored.

1. Deploy NFS server and clients on your Kubernetes cluster. Please make sure they are functional before you go to the next step. Here is a instruction on how to verify NFS service.

    1) On your NFS client node, check if NFS shared dir exists.
    ```
    $ showmount -e 192.168.0.250
    Export list for 192.168.0.250:
    /data/nfs/k8s 192.168.0.0/16
    ``` 
    2) On your NFS client node, create a test dir and mount it to the NFS shared dir (you may need sudo permission).
    ```
    $ mkdir -p   ~/test-nfc
    $ mount -t nfs 192.168.0.250:/data/nfs/k8s ~/test-nfc
    ```
    3) On your NFS client node, create a test file on the mounted test dir.
    ```
    $ touch ~/test-nfc/test.txt
    ```
    4) On your NFS server node, check the shared dir. If there exists the test file we created on the client node, it proves the NFS service is functional.
    ```
    $ ls -ls /data/k8s/
    total 4
    4 -rw-r--r--. 1 root root 4 Jul 10 21:50 test.txt
    ```
2. Modify the following configurations of the `deploy/storage/nfs/nfs-client.yaml` file:
```
...
            - name: NFS_SERVER
              value: 192.168.0.250
            - name: NFS_PATH
              value: /data/nfs/k8s
      volumes:
        - name: nfs-client-root
          nfs:
            server: 192.168.0.250
            path: /data/nfs/k8s
...
```
Replace `192.168.0.250` and `/data/nfs/k8s` with your true NFS server IP address and NFS server data volume path.
3. Create a NFS storage class for RocketMQ, run
```
$ cd deploy/storage/nfs
$ ./deploy-storage-class.sh
```
4. If the storage class is successfully deployed, you can get the pod status like:
```
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
nfs-client-provisioner-7758ff457c-mszgc   1/1     Running   0          136m
rocketmq-operator-6cb8f7d6c4-79m2j        1/1     Running   0          102s
```
### Define Your RocketMQ Cluster

RocketMQ Operator provides several CRDs to allow users define their RocketMQ service component cluster, which includes the Name Server cluster and the Broker cluster.

1. Check the file `rocketmq.zeratullich.org_v1beta1_nameservice_cr.yaml` in the `example` directory, for example:
```
apiVersion: rocketmq.zeratullich.org/v1beta1
kind: NameService
metadata:
  #name: example-nameservice
  name: name-service
spec:
  # Add fields here
  size: 2
  # nameServiceImage is the customized docker image repo of the RocketMQ name service
  nameServiceImage: harbor.myquanwei.com/library/rocketmq-namesrv:4.5.2-alpine 
  # imagePullPolicy is the image pull policy
  imagePullPolicy: Always
  # storageMode can be EmptyDir, HostPath, NFS
  storageMode: NFS
  # hostPath is the local path to store data
  hostPath: /data/rocketmq/nameserver
  # volumeClaimTemplates defines the storageClass
  volumeClaimTemplates:
    - metadata:
        name: namesrv-storage
        annotations:
          volume.beta.kubernetes.io/storage-class: managed-nfs-storage
      spec:
        accessModes: [ "ReadWriteOnce" ]
        resources:
          requests:
            storage: 5Gi
```
### Create RocketMQ Cluster

1. Deploy the RocketMQ name service cluster by running:

``` 
$ kubectl apply -f rocketmq.zeratullich.org_v1beta1_nameservice_cr.yaml
nameservice.rocketmq.zeratullich.org/name-service created
```
Check the status:

```
$ kubectl get svc -owide
NAME                        TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)             AGE   SELECTOR
name-service                ClusterIP   None             <none>        9876/TCP            26h   app=name-service
$ kubectl get pods -owide
NAME                                 READY   STATUS    RESTARTS   AGE    IP             NODE     NOMINATED NODE   READINESS GATES
name-service-0                       1/1     Running   0          163m   10.244.8.246   k8s-06   <none>          <none>
name-service-1                       1/1     Running   0          26h    10.244.2.58    k8s-03   <none>           <none>
rocketmq-operator-6cb8f7d6c4-79m2j   1/1     Running   0          172m   10.244.8.245   k8s-06   <none>           <none>
```
2. Deploy the RocketMQ broker clusters by running:
```
$ kubectl apply -f example/rocketmq_v1alpha1_broker_cr.yaml
broker.rocketmq.zeratullich.org/broker created 
```
After a while the Broker Containers will be created, the Kubernetes clusters status should be like:

``` 
$ kubectl get pods -owide
NAME                                 READY   STATUS    RESTARTS   AGE    IP             NODE     NOMINATED NODE   READINESS GATES
broker-0-master-0                    1/1     Running   0          171m   10.244.2.60    k8s-03   <none>           <none>
broker-0-replica-1-0                 1/1     Running   0          26h    10.244.8.242   k8s-06   <none>           <none>
broker-1-master-0                    1/1     Running   0          26h    10.244.4.195   k8s-07   <none>           <none>
broker-1-replica-1-0                 1/1     Running   0          173m   10.244.6.133   k8s-08   <none>           <none>
broker-2-master-0                    1/1     Running   0          175m   10.244.3.33    k8s-05   <none>           <none>
broker-2-replica-1-0                 1/1     Running   0          175m   10.244.8.244   k8s-06   <none>           <none>
name-service-0                       1/1     Running   0          163m   10.244.8.246   k8s-06   <none>           <none>
name-service-1                       1/1     Running   0          26h    10.244.2.58    k8s-03   <none>           <none>
rocketmq-operator-6cb8f7d6c4-79m2j   1/1     Running   0          172m   10.244.8.245   k8s-06   <none>           <none>
```
3. Check the PV and PVC status:
```
$ kubectl get pvc
NAME                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS          AGE
broker-storage-broker-0-master-0      Bound    pvc-82b9f743-3d35-492e-b02b-461fe4a4de17   15Gi       RWO            managed-nfs-storage   15d
broker-storage-broker-0-replica-1-0   Bound    pvc-d63d9bf1-b360-4e8d-b9ae-d218edbab5b8   15Gi       RWO            managed-nfs-storage   15d
broker-storage-broker-1-master-0      Bound    pvc-f8b7dc0c-37d8-4b54-a588-087465b24fc4   15Gi       RWO            managed-nfs-storage   15d
broker-storage-broker-1-replica-1-0   Bound    pvc-6258e872-7867-4263-8c62-e26e914ad98a   15Gi       RWO            managed-nfs-storage   15d
broker-storage-broker-2-master-0      Bound    pvc-3d938881-d7bc-44c1-acaa-89532fe9ec28   15Gi       RWO            managed-nfs-storage   15d
broker-storage-broker-2-replica-1-0   Bound    pvc-866bde15-4efb-4361-bf5e-40d710be57c2   15Gi       RWO            managed-nfs-storage   15d
namesrv-storage-name-service-0        Bound    pvc-b6cbaa8a-5bc8-4cf4-b32c-cdcd443d60d8   5Gi        RWO            managed-nfs-storage   18d
namesrv-storage-name-service-1        Bound    pvc-b64dc012-f2a6-4230-a4dd-09fad6c68e67   5Gi        RWO            managed-nfs-storage   18d
```
> Notice: if you don't choose the NFS storage mode, then the above PV and PVC won't be created.

Congratulations! You have successfully deployed your RocketMQ cluster by RocketMQ Operator.
## Horizontal Scale
### Name Server Cluster Scale
If the current name service cluster scale does not fit your requirements, you can simply use RocketMQ-Operator to up-scale or down-scale your name service cluster.

If you want to enlarge your name service cluster. Modify your name service CR file `rocketmq.zeratullich.org_v1beta1_nameservice_cr.yaml`, increase the field `size` to the number you want, for example, from `size: 1` to `size: 2`.

> Notice: if your broker image version is 4.5.0 or earlier, you need to make sure that `allowRestart: true` is set in the broker CR file to enable rolling restart policy. If `allowRestart: false`, configure it to `allowRestart: true` and run `kubectl apply -f example/rocketmq.zeratullich.org_v1beta1_broker_cr.yaml` to apply the new config.

After configuring the `size` fields, simply run 
```
kubectl apply -f example/zeratullich.org_v1beta1_nameservice_cr.yaml 
```
Then a new name service pod will be deployed and meanwhile the operator will inform all the brokers to update their name service list parameters, so they can register to the new name service.

> Notice: under the policy `allowRestart: true`, the broker will gradually be updated so the update process is also not perceptible to the producer and consumer clients.

### Broker Cluster Scale
#### Up-scale Broker in Out-of-order Message Scenario
It is often the case that with the development of your business, the old broker cluster scale no longer meets your needs. You can simply use RocketMQ-Operator to up-scale your broker cluster:

1. Modify the `size` in the broker CR file to the number that you want the broker cluster scale will be, for example, from `size: 1` to `size: 2`.
2. Modify the `replicaPerGroup` in the broker CR file to the number that you want the broker slave scale will be, for example, from `replicaPerGroup: 1` to `replicaPerGroup: 2`.
3. Apply the new configurations:
```
kubectl apply -f example/rocketmq.zeratullich.org_v1beta1_broker_cr.yaml
```
Then a new broker group of pods will be deployed and meanwhile the operator will copy the metadata from the source broker pod to the newly created broker pods before the new brokers are stared, so the new brokers will reload previous topic and subscription information.
## Clean the Environment
If you want to tear down the RocketMQ cluster, to remove the broker clusters run
```
$ kubectl delete -f example/rocketmq.zeratullich.org_v1beta1_broker_cr.yaml
```
to remove the name service clusters:
```
$ kubectl delete -f example/rocketmq.zeratullich.org_v1beta1_nameservice_cr.yaml
```
to remove the RocketMQ Operator:
```
$ ./purge-operator.sh
```
to remove the storage class for RocketMQ:
```
$ cd deploy/storage/nfs
$ ./remove-storage-class.sh
```
> Note: the NFS and HostPath persistence data will not be deleted by default.
## Development
### Prerequisites
+ [git](https://git-scm.com/downloads)
+ [go](https://golang.org/dl/) version v1.13+.
+ [docker](https://docs.docker.com/install/) version 17.03+.
+ Access to a Kubernetes v1.15.0+ cluster.
+ [dep](https://golang.github.io/dep/docs/installation.html) version v0.5.0+.
+ [operator-sdk](https://github.com/operator-framework/operator-sdk) version v0.15.2

### Build
For developers who want to build and push the operator-related images to the docker hub, please follow the instructions below.
#### Operator
RocketMQ-Operator uses `operator-sdk` to generate the scaffolding and build the operator image. You can refer to the [operator-sdk user guide](https://github.com/operator-framework/operator-sdk/blob/master/doc/user-guide.md) for more details.

If you want to push the newly build operator image to your own docker hub, please modify the `DOCKERHUB_REPO` variable in the `create-operator.sh` script using your own repository. Then run the build script:
```
$ ./create-operator.sh
```
#### Broker and Name Server Images
RocketMQ-Operator is based on customized images of `Broker` and `Name Server`, which are build by `build-broker-image.sh` and `build-namesrv-image.sh` respectively. Therefore, the images used in the `Broker` and `NameService` CR yaml files should be build by these scripts.

You can also modify the `DOCKERHUB_REPO` variable in the scripts to push the newly build images to your own repository:

```
$ cd images/broker
$ ./build-broker-image.sh
```
```
$ cd images/namesrv
$ ./build-namesrv-image.sh
```
> Note: for users who just want to use the operator, there is no need to build the operator and customized broker and name server images themselves. Users can simply use the default official images which are maintained by the RocketMQ community. 