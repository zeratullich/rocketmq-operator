package nameservice

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	rocketmqv1beta1 "github.com/zeratullich/rocketmq-operator/pkg/apis/rocketmq/v1beta1"
	cons "github.com/zeratullich/rocketmq-operator/pkg/constants"
	"github.com/zeratullich/rocketmq-operator/pkg/controller/share"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nameservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new NameService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNameService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nameservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NameService
	err = c.Watch(&source.Kind{Type: &rocketmqv1beta1.NameService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner NameService
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rocketmqv1beta1.NameService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNameService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNameService{}

// ReconcileNameService reconciles a NameService object
type ReconcileNameService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NameService object and makes changes based on the state read
// and what is in the NameService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNameService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NameService")

	// Fetch the NameService instance
	nameservice := &rocketmqv1beta1.NameService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, nameservice)
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

	// Define a new service object
	svc := r.newService(nameservice)

	// Define a new statefulset object
	dep := r.statefulSetForNameService(nameservice)

	// Check if the statefulSet already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, dep)
	if err != nil && errors.IsNotFound(err) {
		//Create statefulset
		if err := r.client.Create(context.TODO(), dep); err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet of NameService", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			return reconcile.Result{}, err
		}
		// StatefulSet created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get NameService StatefulSet.")
		return reconcile.Result{}, err
	}

	// Check if the service already exists, if not create a new one
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, svc)
	if err != nil && errors.IsNotFound(err) {
		// Create service
		if err := r.client.Create(context.TODO(), svc); err != nil {
			reqLogger.Error(err, "Failed to create new Service of NameService", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get NameService Service.")
		return reconcile.Result{}, err
	}

	// Ensure the statefulSet size is the same as the spec
	size := nameservice.Spec.Size
	if *dep.Spec.Replicas != size {
		dep.Spec.Replicas = &size
		err = r.client.Update(context.TODO(), dep)
		reqLogger.Info("NameService Updated")
		if err != nil {
			reqLogger.Error(err, "Failed to update StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			return reconcile.Result{}, err
		}
	}

	// Ensure the java_opt is the same as the spec
	javaOpt := getJavaOpt(nameservice)
	if dep.Spec.Template.Spec.Containers[0].Env[0].Value != javaOpt {
		dep.Spec.Template.Spec.Containers[0].Env[0].Value = javaOpt
		err = r.client.Update(context.TODO(), dep)
		if err != nil {
			reqLogger.Error(err, "Failed to update JAVA_OPT env of StatefulSet.", "StatefulSet.Namespace", dep.Namespace, "StatefulSet.Name", dep.Name)
			return reconcile.Result{}, err
		}
	}

	return r.updateNameServiceStatus(nameservice, request, true)
}

// newService represents creating service resource
func (r *ReconcileNameService) newService(nameService *rocketmqv1beta1.NameService) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameService.Name,
			Namespace: nameService.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:               nameService.GetName(),
					Kind:               nameService.Kind,
					APIVersion:         nameService.APIVersion,
					UID:                nameService.GetUID(),
					Controller:         &(share.BoolTrue),
					BlockOwnerDeletion: &(share.BoolTrue),
				},
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "cluster",
					Port:       cons.NameServiceMainContainerPort,
					TargetPort: intstr.FromInt(cons.NameServiceMainContainerPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": nameService.Name,
			},
		},
	}
}

func (r *ReconcileNameService) updateNameServiceStatus(nameService *rocketmqv1beta1.NameService, request reconcile.Request, requeue bool) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Check the NameServers status")
	// List the pods for this nameService's statefulSet
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForNameService(nameService.Name))
	listOps := &client.ListOptions{
		Namespace:     nameService.Namespace,
		LabelSelector: labelSelector,
	}

	err := r.client.List(context.TODO(), podList, listOps)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods.", "NameService.Namespace", nameService.Namespace, "NameService.Name", nameService.Name)
		return reconcile.Result{Requeue: true}, err
	}

	serverNames := getNameServers(nameService.Name, nameService.Namespace, nameService.Spec.Size)
	nameServerListStr := ""
	for _, value := range serverNames {
		var builder strings.Builder
		builder.WriteString(nameServerListStr)
		builder.WriteString(value)
		builder.WriteString(":9876;")
		nameServerListStr = builder.String()
	}
	if nameServerListStr == "" {
		share.NameServersStr = nameServerListStr
	} else {
		share.NameServersStr = nameServerListStr[:len(nameServerListStr)-1]
	}

	// Update status.NameServers if needed
	if !reflect.DeepEqual(serverNames, nameService.Status.NameServers) {
		reqLogger.Info("share.NameServersStr:" + share.NameServersStr)
		oldNameServerListStr := ""
		for _, value := range nameService.Status.NameServers {
			var builder strings.Builder
			builder.WriteString(oldNameServerListStr)
			builder.WriteString(value)
			builder.WriteString(":9876;")
			oldNameServerListStr = builder.String()
		}

		if len(oldNameServerListStr) == 0 {
			oldNameServerListStr = share.NameServersStr
		} else {
			oldNameServerListStr = oldNameServerListStr[:len(oldNameServerListStr)-1]
		}

		reqLogger.Info("oldNameServerListStr:" + oldNameServerListStr)

		nameService.Status.NameServers = serverNames
		err := r.client.Status().Update(context.TODO(), nameService)
		// Update the NameServers status with the host name
		reqLogger.Info("Updated the NameServers status with the host name")
		if err != nil {
			reqLogger.Error(err, "Failed to update NameServers status of NameService.")
			return reconcile.Result{Requeue: true}, err
		}
	}

	// Print NameServers
	for i, value := range nameService.Status.NameServers {
		reqLogger.Info("NameServers Name " + strconv.Itoa(i) + ": " + value)
	}

	if requeue {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func getJavaOpt(nameService *rocketmqv1beta1.NameService) string {
	xmx := nameService.Spec.Xmx
	xms := nameService.Spec.Xms
	xmn := nameService.Spec.Xmn
	return fmt.Sprintf("-Xmx%s -Xmn%s -Xmn%s", xmx, xmn, xms)
}

func getVolumeClaimTemplates(nameService *rocketmqv1beta1.NameService) []corev1.PersistentVolumeClaim {
	switch nameService.Spec.StorageMode {
	case cons.StorageModeNFS:
		return nameService.Spec.VolumeClaimTemplates
	case cons.StorageModeEmptyDir, cons.StorageModeHostPath:
		fallthrough
	default:
		return nil
	}
}

func getVolumes(nameService *rocketmqv1beta1.NameService) []corev1.Volume {
	switch nameService.Spec.StorageMode {
	case cons.StorageModeNFS:
		return nil
	case cons.StorageModeEmptyDir:
		volumes := []corev1.Volume{{
			Name: nameService.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{}},
		}}
		return volumes
	case cons.StorageModeHostPath:
		fallthrough
	default:
		volumes := []corev1.Volume{{
			Name: nameService.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: nameService.Spec.HostPath,
				}},
		}}
		return volumes
	}
}

func getNameServers(name string, namespace string, size int32) []string {
	var times int32
	var nameServers []string
	for {
		var build strings.Builder
		if times >= size {
			break
		}
		build.WriteString(name)
		build.WriteString("-")
		build.WriteString(strconv.FormatInt(int64(times), 10))
		build.WriteString(".")
		build.WriteString(name)
		build.WriteString(".")
		build.WriteString(namespace)
		build.WriteString(".svc.cluster.local")
		nameServers = append(nameServers, build.String())
		times++
	}
	return nameServers
}

func labelsForNameService(name string) map[string]string {
	return map[string]string{"app": name, "name_service_cr": name}
}

func (r *ReconcileNameService) statefulSetForNameService(nameService *rocketmqv1beta1.NameService) *appsv1.StatefulSet {
	ls := labelsForNameService(nameService.Name)
	var mountPathName string

	if getVolumeClaimTemplates(nameService) != nil {
		mountPathName = nameService.Spec.VolumeClaimTemplates[0].Name
	} else {
		mountPathName = nameService.Name
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameService.Name,
			Namespace: nameService.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					Name:               nameService.GetName(),
					Kind:               nameService.Kind,
					APIVersion:         nameService.APIVersion,
					UID:                nameService.GetUID(),
					Controller:         &(share.BoolTrue),
					BlockOwnerDeletion: &(share.BoolTrue),
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &nameService.Spec.Size,
			ServiceName: nameService.Name,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					DNSPolicy: "ClusterFirst",
					Affinity:  &nameService.Spec.Affinity,
					Containers: []corev1.Container{{
						Image:           nameService.Spec.NameServiceImage,
						Name:            nameService.Name,
						ImagePullPolicy: nameService.Spec.ImagePullPolicy,
						Ports: []corev1.ContainerPort{{
							ContainerPort: cons.NameServiceMainContainerPort,
							Name:          cons.NameServiceMainContainerPortName,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: cons.LogMountPath,
							Name:      mountPathName,
							SubPath:   cons.LogSubPathName,
						}},
						Env: []corev1.EnvVar{{
							Name:  cons.JavaOpt,
							Value: getJavaOpt(nameService),
						}},
						Resources: nameService.Spec.Resources,
					}},
					Volumes: getVolumes(nameService),
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplates(nameService),
		},
	}
	return sts
}
