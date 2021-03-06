package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NameServiceSpec defines the desired state of NameService
// +k8s:openapi-gen=true
type NameServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// Size is the number of the name service Pod
	Size int32 `json:"size"`
	//NameServiceImage is the name service image
	NameServiceImage string `json:"nameServiceImage"`
	// ImagePullPolicy defines how the image is pulled.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
	// StorageMode can be EmptyDir, HostPath, StorageClass
	StorageMode string `json:"storageMode"`
	// HostPath is the local path to store data
	HostPath string `json:"hostPath"`
	// VolumeClaimTemplates defines the StorageClass
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	// Resources limits pod resource usage
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Java memory limits Xmx
	Xmx string `json:"xmx"`
	// Java memory limits Xms
	Xms string `json:"xms"`
	// Java memory limits Xmn
	Xmn string `json:"xmn"`
	// Affinity is a group of affinity scheduling rules.
	Affinity corev1.Affinity `json:"affinity,omitempty"`
}

// NameServiceStatus defines the observed state of NameService
// +k8s:openapi-gen=true
type NameServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	// NameServers is the name service ip list
	NameServers []string `json:"nameServers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NameService is the Schema for the nameservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type NameService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NameServiceSpec   `json:"spec,omitempty"`
	Status NameServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NameServiceList contains a list of NameService
type NameServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NameService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NameService{}, &NameServiceList{})
}
