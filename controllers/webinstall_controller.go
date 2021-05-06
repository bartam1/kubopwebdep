/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/bartam1/kubopwebdep/api/v1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// WebInstallReconciler reconciles a WebInstall object
type WebInstallReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=crd.bartam,resources=webinstalls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.bartam,resources=webinstalls/status,verbs=get;update;patch

func (r *WebInstallReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	ctx := context.Background()
	logger := r.Log.WithValues("namespace", req.NamespacedName)
	logger.Info("Reconciling...")
	webInstall := &v1.WebInstall{}
	err := r.Get(ctx, req.NamespacedName, webInstall)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request - return and don't requeue:
			logger.Info("request object and deployment deleted", "RO name", req.Name)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request:
		return reconcile.Result{}, err
	}
	logger.WithValues("Status: ", webInstall.Status, "Image: ", webInstall.Spec.Image, "Host: ", webInstall.Spec.Host, "Replicas: ", webInstall.Spec.Replicas).Info("")

	//Remove any other deployments (other hosts)

	logger.Info("checking if an existing Deployment exists for this resource")
	deployment := &apps.Deployment{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name + "-" + webInstall.Spec.Host}, deployment)
	if apierrors.IsNotFound(err) {
		logger.Info("could not found existing Deployment for this RO, creating one...")
		deployment = buildDeployment(*webInstall)
		//Bind webInstall to that deployment for reconcile at changes
		if err := controllerutil.SetControllerReference(webInstall, deployment, r.Scheme); err != nil {
			// requeue with error
			return reconcile.Result{}, err
		}
		if err := r.Client.Create(ctx, deployment); err != nil {
			logger.Error(err, "failed to create Deployment resource")
			return ctrl.Result{}, err
		}
		//r.Recorder.Eventf(webInstall, core.EventTypeNormal, "Created", "Created deployment %q", deployment.Name)
		logger.Info("deployment resource created for that RO")
		webInstall.Status.Phase = v1.PhaseRunning
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "failed to get Deployment for that RO")
		return ctrl.Result{}, err
	}

	logger.Info("existing Deployment resource already exists for that RO, checking replica count")
	if deployment.Status.UnavailableReplicas > 0 {
		logger.Info("unavailable replica(s) found", "count", deployment.Status.UnavailableReplicas)
	}

	if *deployment.Spec.Replicas != webInstall.Spec.Replicas {
		logger.Info("different number of replicas --> updating", "old_count", *deployment.Spec.Replicas, "new_count", webInstall.Spec.Replicas)

		deployment.Spec.Replicas = &webInstall.Spec.Replicas
		if err := r.Client.Update(ctx, deployment); err != nil {
			logger.Error(err, "failed to Deployment update replica count")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}
	logger.Info("replica count is OK", "current", *deployment.Spec.Replicas)
	return reconcile.Result{}, err
}

func (r *WebInstallReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.WebInstall{}).
		Owns(&apps.Deployment{}).
		Complete(r)
}
