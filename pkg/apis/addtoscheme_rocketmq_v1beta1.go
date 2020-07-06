package apis

import (
	"github.com/zeratullich/rocketmq-operator/pkg/apis/rocketmq/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1beta1.SchemeBuilder.AddToScheme)
}
