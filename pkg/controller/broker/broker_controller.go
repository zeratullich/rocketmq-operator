package broker

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"time"

	rocketmqv1beta1 "github.com/project/rocketmq-operator/pkg/apis/rocketmq/v1beta1"
	cons "github.com/project/rocketmq-operator/pkg/constants"
	"github.com/project/rocketmq-operator/pkg/controller/share"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_broker")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Broker Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBroker{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("broker-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Broker
	err = c.Watch(&source.Kind{Type: &rocketmqv1beta1.Broker{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Broker
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1beta1.Broker{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileBroker implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBroker{}

// ReconcileBroker reconciles a Broker object
type ReconcileBroker struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Broker object and makes changes based on the state read
// and what is in the Broker.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBroker) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Broker")

	// Fetch the Broker instance
	broker := &rocketmqv1beta1.Broker{}
	err := r.client.Get(context.TODO(), request.NamespacedName, broker)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	var isInitial = true
	var groupNum int

	if isInitial || broker.Status.Size == 0 {
		groupNum = broker.Spec.Size
	} else {
		groupNum = broker.Status.Size
	}

	replicaPerGroup := broker.Spec.ReplicaPerGroup

	reqLogger.Info("brokerGroupNum=" + strconv.Itoa(groupNum) + ", replicaPerGroup=" + strconv.Itoa(replicaPerGroup))
	for brokerGroupIndex := 0; brokerGroupIndex < groupNum; brokerGroupIndex++ {
		reqLogger.Info("Check Broker cluster " + strconv.Itoa(brokerGroupIndex+1) + "/" + strconv.Itoa(groupNum))
		masterDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, 0)
		// Check if the master broker statefulSet already exists, if not create a new one
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: masterDep.Name, Namespace: masterDep.Namespace}, masterDep)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Master Broker StatefulSet.", "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
			err = r.client.Create(context.TODO(), masterDep)
			if err != nil {
				reqLogger.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
				return reconcile.Result{}, err
			}
		} else if err != nil {
			reqLogger.Error(err, "Failed to get broker master StatefulSet.")
			return reconcile.Result{}, err
		}

		for replicaIndex := 1; replicaIndex <= replicaPerGroup; replicaIndex++ {
			reqLogger.Info("Check Replica Broker of cluster-" + strconv.Itoa(brokerGroupIndex) + " " + strconv.Itoa(replicaIndex) + "/" + strconv.Itoa(replicaPerGroup))
			replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, replicaIndex)
			// Check if the replica broker statefulSet already exists, if not create a new one
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, replicaDep)
			if err != nil && errors.IsNotFound(err) {
				reqLogger.Info("Creating a new Replica Broker StatefulSet.", "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
				err = r.client.Create(context.TODO(), replicaDep)
				if err != nil {
					reqLogger.Error(err, "Failed to create new StatefulSet of broker-"+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
					return reconcile.Result{Requeue: true}, err
				}
			} else if err != nil {
				reqLogger.Error(err, "Failed to get broker replica StatefulSet.")
				return reconcile.Result{Requeue: true}, err
			}
		}
	}

	// Check for name server scaling
	if broker.Spec.AllowRestart {
		// The following code will  restart all brokers to update NAMESRV_ADDR env if share.NameServersStr is changed
		for brokerGroupIndex := 0; brokerGroupIndex < broker.Spec.Size; brokerGroupIndex++ {
			brokerName := getBrokerName(broker, brokerGroupIndex)
			// Update master broker
			masterDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, 0)
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: masterDep.Name, Namespace: masterDep.Namespace}, masterDep)
			if err != nil {
				reqLogger.Error(err, "Failed to get broker master StatefulSet of "+brokerName)
			} else {
				if masterDep.Spec.Template.Spec.Containers[0].Env[0].Value != share.NameServersStr {
					masterDep.Spec.Template.Spec.Containers[0].Env[0].Value = share.NameServersStr
					err = r.client.Update(context.TODO(), masterDep)
					if err != nil {
						reqLogger.Error(err, "Failed to update NAMESRV_ADDR of master broker "+brokerName, "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
					} else {
						reqLogger.Info("Successfully updated NAMESRV_ADDR of master broker "+brokerName, "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
					}
					time.Sleep(time.Duration(cons.RestartBrokerPodIntervalInSecond) * time.Second)
				}
			}

			// Update replicas brokers
			for replicaIndex := 1; replicaIndex <= replicaPerGroup; replicaIndex++ {
				replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, replicaIndex)
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, replicaDep)
				if err != nil {
					reqLogger.Error(err, "Failed to get broker replica StatefulSet of "+brokerName)
				} else {
					if replicaDep.Spec.Template.Spec.Containers[0].Env[0].Value != share.NameServersStr {
						replicaDep.Spec.Template.Spec.Containers[0].Env[0].Value = share.NameServersStr
						err = r.client.Update(context.TODO(), replicaDep)
						if err != nil {
							reqLogger.Error(err, "Failed to update NAMESRV_ADDR of "+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
						} else {
							reqLogger.Info("Successfully updated NAMESRV_ADDR of "+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
						}
						time.Sleep(time.Duration(cons.RestartBrokerPodIntervalInSecond) * time.Second)
					}
				}
			}
		}
	}

	// List the pods for this broker's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForBroker(broker.Name))
	listOps := &client.ListOptions{
		Namespace:     broker.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), podList, listOps)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "Broker.Namespace", broker.Namespace, "Broker.Name", broker.Name)
		return reconcile.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	log.Info("broker.Status.Nodes length = " + strconv.Itoa(len(broker.Status.Nodes)))
	log.Info("podNames length = " + strconv.Itoa(len(podNames)))

	// Ensure every pod is in running phase, then change the isInitial state to false
	notReady := false

	for _, pod := range podList.Items {
		if !reflect.DeepEqual(pod.Status.Phase, corev1.PodRunning) {
			log.Info("pod " + pod.Name + " phase is " + string(pod.Status.Phase) + ", wait for a moment...")
			notReady = true
		}
	}
	if !notReady {
		isInitial = false
	}

	// Update status.Size if needed
	if broker.Spec.Size > broker.Status.Size {
		log.Info("broker.Status.Size = " + strconv.Itoa(broker.Status.Size))
		log.Info("broker.Spec.Size = " + strconv.Itoa(broker.Spec.Size))
		broker.Status.Size = broker.Spec.Size
		err = r.client.Status().Update(context.TODO(), broker)
		if err != nil {
			reqLogger.Error(err, "Failed to update Broker Master Size status.")
		}
	} else if broker.Spec.Size < broker.Status.Size {
		log.Info("broker.Status.Size = " + strconv.Itoa(broker.Status.Size))
		log.Info("broker.Spec.Size = " + strconv.Itoa(broker.Spec.Size))
		for brokerGroupIndex := broker.Status.Size; brokerGroupIndex > broker.Spec.Size; brokerGroupIndex-- {
			masterDep := r.getBrokerStatefulSet(broker, brokerGroupIndex-1, 0)
			// Delete extra master broker
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: masterDep.Name, Namespace: masterDep.Namespace}, masterDep)
			if err == nil {
				reqLogger.Info("Delete Master Broker StatefulSet.", "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
				err = r.client.Delete(context.TODO(), masterDep)
				if err != nil {
					reqLogger.Error(err, "Failed to delete StatefulSet", "StatefulSet.Namespace", masterDep.Namespace, "StatefulSet.Name", masterDep.Name)
				}
			} else {
				reqLogger.Error(err, "Failed to get broker master StatefulSet.")
			}
			// Delete extra replicas brokers
			for replicaIndex := 1; replicaIndex <= replicaPerGroup; replicaIndex++ {
				replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex-1, replicaIndex)
				err = r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, masterDep)
				if err == nil {
					reqLogger.Info("Delete Replica Broker StatefulSet.", "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
					err = r.client.Delete(context.TODO(), replicaDep)
					if err != nil {
						reqLogger.Error(err, "Failed to delete StatefulSet of broker-"+strconv.Itoa(brokerGroupIndex-1)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
					}
				} else {
					reqLogger.Error(err, "Failed to get broker replica StatefulSet.")
				}
			}
		}

		broker.Status.Size = broker.Spec.Size
		err = r.client.Status().Update(context.TODO(), broker)
		if err != nil {
			reqLogger.Error(err, "Failed to update Broker Size status.")
		}
	}

	// Update status.replicaPerGroup if needed
	if broker.Status.ReplicaPerGroup < replicaPerGroup {
		broker.Status.ReplicaPerGroup = replicaPerGroup
		if err := r.client.Status().Update(context.TODO(), broker); err != nil {
			reqLogger.Error(err, "Failed to update Broker Replicas Size status.")
		}
	} else if broker.Status.ReplicaPerGroup > replicaPerGroup {
		// Delete extra replicas brokers
		for brokerGroupIndex := 0; brokerGroupIndex < broker.Spec.Size; brokerGroupIndex++ {
			for replicaIndex := broker.Status.ReplicaPerGroup; replicaIndex > replicaPerGroup; replicaIndex-- {
				replicaDep := r.getBrokerStatefulSet(broker, brokerGroupIndex, replicaIndex)
				if err := r.client.Get(context.TODO(), types.NamespacedName{Name: replicaDep.Name, Namespace: replicaDep.Namespace}, replicaDep); err == nil {
					reqLogger.Info("Delete Replica Broker StatefulSet.", "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
					if err = r.client.Delete(context.TODO(), replicaDep); err != nil {
						reqLogger.Error(err, "Failed to delete StatefulSet of broker-"+strconv.Itoa(brokerGroupIndex)+"-replica-"+strconv.Itoa(replicaIndex), "StatefulSet.Namespace", replicaDep.Namespace, "StatefulSet.Name", replicaDep.Name)
					}
				} else {
					reqLogger.Error(err, "Failed to get broker replica StatefulSet.")
				}
			}
		}
		broker.Status.ReplicaPerGroup = replicaPerGroup
		err = r.client.Status().Update(context.TODO(), broker)
		if err != nil {
			reqLogger.Error(err, "Failed to update Broker ReplicaPerGroup status.")
		}
	}

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, broker.Status.Nodes) {
		broker.Status.Nodes = podNames
		err = r.client.Status().Update(context.TODO(), broker)
		if err != nil {
			reqLogger.Error(err, "Failed to update Broker Nodes status.")
		}
	}

	return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
}

func getBrokerName(broker *rocketmqv1beta1.Broker, brokerGroupIndex int) string {
	var builder strings.Builder
	builder.WriteString(broker.Name)
	builder.WriteString("-")
	builder.WriteString(strconv.Itoa(brokerGroupIndex))
	return builder.String()
}

// getBrokerStatefulSet returns a broker StatefulSet object
func (r *ReconcileBroker) getBrokerStatefulSet(broker *rocketmqv1beta1.Broker, brokerGroupIndex int, replicaIndex int) *appsv1.StatefulSet {
	var builder strings.Builder
	ls := labelsForBroker(broker.Name)
	var a int32 = 1
	var c = &a
	var statefulSetName string
	var nameServers = share.NameServersStr
	if replicaIndex == 0 {
		builder.WriteString(broker.Name)
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(brokerGroupIndex))
		builder.WriteString("-master")
	} else {
		builder.WriteString(broker.Name)
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(brokerGroupIndex))
		builder.WriteString("-replica-")
		builder.WriteString(strconv.Itoa(replicaIndex))
	}
	statefulSetName = builder.String()

	dep := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      statefulSetName,
			Namespace: broker.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:               broker.GetName(),
					Kind:               broker.Kind,
					APIVersion:         broker.APIVersion,
					UID:                broker.GetUID(),
					Controller:         &(share.BoolTrue),
					BlockOwnerDeletion: &(share.BoolTrue),
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: c,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           broker.Spec.BrokerImage,
						Name:            cons.BrokerContainerName,
						ImagePullPolicy: broker.Spec.ImagePullPolicy,
						Env: []corev1.EnvVar{{
							Name:  cons.EnvNameServiceAddress,
							Value: nameServers,
						}, {
							Name:  cons.EnvReplicationMode,
							Value: broker.Spec.ReplicationMode,
						}, {
							Name:  cons.EnvBrokerID,
							Value: strconv.Itoa(replicaIndex),
						}, {
							Name:  cons.EnvBrokerClusterName,
							Value: broker.Name + "-" + strconv.Itoa(brokerGroupIndex),
						}, {
							Name:  cons.EnvBrokerName,
							Value: broker.Name + "-" + strconv.Itoa(brokerGroupIndex),
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: cons.BrokerVipContainerPort,
							Name:          cons.BrokerVipContainerPortName,
						}, {
							ContainerPort: cons.BrokerMainContainerPort,
							Name:          cons.BrokerMainContainerPortName,
						}, {
							ContainerPort: cons.BrokerHighAvailabilityContainerPort,
							Name:          cons.BrokerHighAvailabilityContainerPortName,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name:      broker.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.LogSubPathName + getPathSuffix(broker, brokerGroupIndex, replicaIndex),
						}, {
							MountPath: cons.StoreMountPath,
							Name:      broker.Spec.VolumeClaimTemplates[0].Name,
							SubPath:   cons.StoreSubPathName + getPathSuffix(broker, brokerGroupIndex, replicaIndex),
						}},
					}},
					Volumes: getVolumes(broker),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplates(broker),
		},
	}

	return dep

}

func getVolumeClaimTemplates(broker *rocketmqv1beta1.Broker) []corev1.PersistentVolumeClaim {
	switch broker.Spec.StorageMode {
	case cons.StorageModeNFS:
		return broker.Spec.VolumeClaimTemplates
	case cons.StorageModeEmptyDir, cons.StorageModeHostPath:
		fallthrough
	default:
		return nil
	}
}

func getVolumes(broker *rocketmqv1beta1.Broker) []corev1.Volume {
	switch broker.Spec.StorageMode {
	case cons.StorageModeNFS:
		return nil
	case cons.StorageModeEmptyDir:
		volumes := []corev1.Volume{{
			Name: broker.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{}},
		}}
		return volumes
	case cons.StorageModeHostPath:
		fallthrough
	default:
		volumes := []corev1.Volume{{
			Name: broker.Spec.VolumeClaimTemplates[0].Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: broker.Spec.HostPath,
				}},
		}}
		return volumes
	}
}

func getPathSuffix(broker *rocketmqv1beta1.Broker, brokerGroupIndex int, replicaIndex int) string {
	var builder strings.Builder
	if replicaIndex == 0 {
		builder.WriteString("/")
		builder.WriteString(broker.Name)
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(brokerGroupIndex))
		builder.WriteString("-master")
	} else {
		builder.WriteString("/")
		builder.WriteString(broker.Name)
		builder.WriteString("-")
		builder.WriteString(strconv.Itoa(brokerGroupIndex))
		builder.WriteString("-replica-")
		builder.WriteString(strconv.Itoa(replicaIndex))
	}
	return builder.String()
}

// labelsForBroker returns the labels for selecting the resources
// belonging to the given broker CR name.
func labelsForBroker(name string) map[string]string {
	return map[string]string{"app": "broker", "broker_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
