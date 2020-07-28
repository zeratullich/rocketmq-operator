package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BrokerSpec defines the desired state of Broker
// +k8s:openapi-gen=true
type BrokerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Size is the number of the broker service Pod
	Size int `json:"size"`
	// ReplicationMode is SYNC or ASYNC
	ReplicationMode string `json:"replicationMode,omitempty"`
	// ReplicaPerGroup each broker cluster's replica number
	ReplicaPerGroup int `json:"replicaPerGroup"`
	// BaseImage is the broker image to use for the Pods.
	BrokerImage string `json:"brokerImage"`
	// ImagePullPolicy defines how the image is pulled.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	// AllowRestart defines whether allow pod restart
	AllowRestart bool `json:"allowRestart"`
	// StorageMode can be EmptyDir, HostPath, NFS
	StorageMode string `json:"storageMode"`
	// HostPath is the local path to store data
	HostPath string `json:"hostPath"`
	// Resources limits pod resource usage
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// VolumeClaimTemplates defines the StorageClass
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	// Java memory limits Xmx
	Xmx string `json:"xmx"`
	// Java memory limits Xms
	Xms string `json:"xms"`
	// Java memory limits Xmn
	Xmn string `json:"xmn"`
	// BrokerClusterName is the cluster name of brokers
	BrokerClusterName string `json:"brokerClusterName"`
	// Affinity is a group of affinity scheduling rules.
	Affinity corev1.Affinity `json:"affinity,omitempty"`
}

// BrokerStatus defines the observed state of Broker
// +k8s:openapi-gen=true
type BrokerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes           []string `json:"nodes,omitempty"`
	Size            int      `json:"size"`
	ReplicaPerGroup int      `json:"replicaPerGroup"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Broker is the Schema for the brokers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=brokers,scope=Namespaced
// +k8s:openapi-gen=true
type Broker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec,omitempty"`
	Status BrokerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BrokerList contains a list of Broker
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broker{}, &BrokerList{})
}
