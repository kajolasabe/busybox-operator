package busybox



import (

	"context"

	"time"



	busyboxv1alpha1 "persistent.com/busybox/busybox-operator/pkg/apis/busybox/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/intstr"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

)



const busyboxPort = 81

//const busyboxNodePort = 80

const busyboxImage = "busybox:1.31.1"



func busyboxDeploymentName(v *busyboxv1alpha1.Busybox) string {

	return v.Name + "-deployment"

}



func busyboxServiceName(v *busyboxv1alpha1.Busybox) string {

	return v.Name + "-service"

}



func (r *ReconcileBusybox) busyboxDeployment(v *busyboxv1alpha1.Busybox) *appsv1.Deployment {

	labels := labels(v, "busybox")

	size := v.Spec.Size

	dep := &appsv1.Deployment{

		ObjectMeta: metav1.ObjectMeta{

			Name:		busyboxDeploymentName(v),

			Namespace: 	v.Namespace,

		},

		Spec: appsv1.DeploymentSpec{

			Replicas: &size,

			Selector: &metav1.LabelSelector{

				MatchLabels: labels,

			},

			Template: corev1.PodTemplateSpec{

				ObjectMeta: metav1.ObjectMeta{

					Labels: labels,

				},

				Spec: corev1.PodSpec{

					Containers: []corev1.Container{{

						Image:	busyboxImage,

						ImagePullPolicy: corev1.PullAlways,

						Name:	"busybox-service",

						Ports:	[]corev1.ContainerPort{{

							ContainerPort: 	busyboxPort,

							Name:			"busybox",

						}},

					}},

				},

			},

		},

	}



	controllerutil.SetControllerReference(v, dep, r.scheme)

	return dep

}



func (r *ReconcileBusybox) busyboxService(v *busyboxv1alpha1.Busybox) *corev1.Service {

	labels := labels(v, "busybox")



	s := &corev1.Service{

		ObjectMeta: metav1.ObjectMeta{

			Name:		busyboxServiceName(v),

			Namespace: 	v.Namespace,

		},

		Spec: corev1.ServiceSpec{

			Selector: labels,

			Ports: []corev1.ServicePort{{

				Protocol: corev1.ProtocolTCP,

				Port: busyboxPort,

				TargetPort: intstr.FromInt(busyboxPort),

				NodePort: 30685,

			}},

			Type: corev1.ServiceTypeNodePort,

		},

	}



	controllerutil.SetControllerReference(v, s, r.scheme)

	return s

}



func (r *ReconcileBusybox) updateBusyboxStatus(v *busyboxv1alpha1.Busybox) (error) {

	//v.Status.BackendImage = busyboxImage

	err := r.client.Status().Update(context.TODO(), v)

	return err

}



func (r *ReconcileBusybox) handleBusyboxChanges(v *busyboxv1alpha1.Busybox) (*reconcile.Result, error) {

	found := &appsv1.Deployment{}

	err := r.client.Get(context.TODO(), types.NamespacedName{

		Name:      busyboxDeploymentName(v),

		Namespace: v.Namespace,

	}, found)

	if err != nil {

		// The deployment may not have been created yet, so requeue

		return &reconcile.Result{RequeueAfter:5 * time.Second}, err

	}



	size := v.Spec.Size



	if size != *found.Spec.Replicas {

		found.Spec.Replicas = &size

		err = r.client.Update(context.TODO(), found)

		if err != nil {

			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)

			return &reconcile.Result{}, err

		}

		// Spec updated - return and requeue

		return &reconcile.Result{Requeue: true}, nil

	}

	return nil, nil

}
