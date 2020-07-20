package constants

const (
	// JavaOpt is the optionals of java
	JavaOpt = "JAVA_OPT"

	// BrokerContainerName is the name of broker container
	BrokerContainerName = "broker"

	// EnvNameServiceAddress is the container environment variable name of name server list
	EnvNameServiceAddress = "NAMESRV_ADDR"

	// EnvReplicationMode is the container environment variable name of replication mode
	EnvReplicationMode = "REPLICATION_MODE"

	// EnvBrokerID is the container environment variable name of broker id
	EnvBrokerID = "BROKER_ID"

	// EnvBrokerClusterName is the container environment variable name of broker cluster name
	EnvBrokerClusterName = "BROKER_CLUSTER_NAME"

	// EnvBrokerName is the container environment variable name of broker name
	EnvBrokerName = "BROKER_NAME"

	// LogMountPath is the directory of RocketMQ log files
	LogMountPath = "/home/rocketmq/logs"

	// StoreMountPath is the directory of RocketMQ store files
	StoreMountPath = "/home/rocketmq/store"

	// LogSubPathName is the sub-path name of log dir under mounted host dir
	LogSubPathName = "logs"

	// StoreSubPathName is the sub-path name of store dir under mounted host dir
	StoreSubPathName = "store"

	// NameServiceMainContainerPort is the main port number of name server container
	NameServiceMainContainerPort = 9876

	// NameServiceMainContainerPortName is the main port name of name server container
	NameServiceMainContainerPortName = "main"

	// BrokerVipContainerPort is the VIP port number of broker container
	BrokerVipContainerPort = 10909

	// BrokerVipContainerPortName is the VIP port name of broker container
	BrokerVipContainerPortName = "vip"

	// BrokerMainContainerPort is the main port number of broker container
	BrokerMainContainerPort = 10911

	// BrokerMainContainerPortName is the main port name of broker container
	BrokerMainContainerPortName = "main"

	// BrokerHighAvailabilityContainerPort is the high availability port number of broker container
	BrokerHighAvailabilityContainerPort = 10912

	// BrokerHighAvailabilityContainerPortName is the high availability port name of broker container
	BrokerHighAvailabilityContainerPortName = "ha"

	// BrokerClusterName is the cluster name of brokers
	BrokerClusterName = "K8S-RocketMQ-Cluster"

	// StorageModeNFS is the name of NFS storage mode
	StorageModeNFS = "NFS"

	// StorageModeEmptyDir is the name of EmptyDir storage mode
	StorageModeEmptyDir = "EmptyDir"

	// StorageModeHostPath is the name pf HostPath storage mode
	StorageModeHostPath = "HostPath"

	// RestartBrokerPodIntervalInSecond is restart broker pod interval in second
	RestartBrokerPodIntervalInSecond = 30

	// RequeueIntervalInSecond is an universal interval of the reconcile function
	RequeueIntervalInSecond = 15
)
