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

	"github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1 "github.com/bartam1/kubopwebdep/api/v1"
)

// WebInstallReconciler reconciles a WebInstall object
type WebInstallReconciler struct {
	client.Client
	Log      *logrus.Entry
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=crd.bartam,resources=webinstalls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crd.bartam,resources=webinstalls/status,verbs=get;update;patch

func (r *WebInstallReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logrus.SetLevel(logrus.InfoLevel)
	ctx := context.Background()

	//r.Log = logrusr.NewLogger(logrus.New())
	//logger := r.Log.WithValues("name", req.Name)
	logger := logrus.WithFields(logrus.Fields{
		"name": req.Name,
	})
	webInstall := &v1.WebInstall{}
	err := r.Get(ctx, req.NamespacedName, webInstall)

	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request - return and don't requeue:
			logger.Warn("request object, deployment, service, ingress has been removed")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request:
		return reconcile.Result{}, err
	}
	//Set new logging fields
	logger = logrus.WithFields(logrus.Fields{
		"0_Name":     req.Name,
		"1_Replicas": webInstall.Spec.Replicas,
		"2_Host":     webInstall.Spec.Host,
		"3_Image":    webInstall.Spec.Image,
	})

	logger.Debug("checking if an Deployment exists for this resource")

	deployment := &apps.Deployment{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, deployment)

	if err != nil && apierrors.IsNotFound(err) {
		webInstall.Status.Phase = v1.PhasePending
		logger.Debug("could not found existing Deployment for this RO")
	} else if err != nil {

		logger.Error(err, "failed to get Deployment for that RO")
		return ctrl.Result{}, err
	} else {
		webInstall.Status.Phase = v1.PhaseRunning
		logger.Debug("existing Deployment resource already exists for that RO")
	}

	switch webInstall.Status.Phase {
	case v1.PhasePending:
		logger.Info("CREATING WEBINSTALL")
		if err := r.createWebInstallBundle(ctx, webInstall, logger); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("WEBINSTALL SUCESSFULLY CREATED")
	case v1.PhaseRunning:
		logger.Info("CHECKING INVARIANT")
		if err = r.updateDeploymentImage(ctx, deployment, webInstall, logger); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.updateReplicas(ctx, deployment, webInstall, logger); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.updateService(ctx, webInstall, logger); err != nil {
			return ctrl.Result{}, err
		}
		if err = r.updateIngress(ctx, webInstall, logger); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("INVARIANT OK")
	}
	return ctrl.Result{}, nil
}

func (r *WebInstallReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.WebInstall{}).
		Owns(&apps.Deployment{}).
		Owns(&core.Service{}).
		Owns(&net.Ingress{}).
		Complete(r)
}
