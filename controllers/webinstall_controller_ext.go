package controllers

import (
	"context"
	"strings"

	v1 "github.com/bartam1/kubopwebdep/api/v1"
	"github.com/sirupsen/logrus"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *WebInstallReconciler) createWebInstallBundle(ctx context.Context, webInstall *v1.WebInstall) (err error) {
	deployment := buildDeployment(webInstall)
	//Bind webInstall to that deployment for reconcile at change and for auto delete
	if err := controllerutil.SetControllerReference(webInstall, deployment, r.Scheme); err != nil {
		return err
	}
	r.Log.Info("deployment resource OK")
	if err := r.Create(ctx, deployment); err != nil {
		r.Log.Error(err, "failed to create deployment resource")
		return err
	}
	service := buildService(webInstall)
	if err := controllerutil.SetControllerReference(webInstall, service, r.Scheme); err != nil {
		return err
	}
	if err = r.Create(ctx, service); err != nil {
		r.Log.Error(err, "failed to create service resource")
		return err
	}
	r.Log.Info("service resource OK")
	ingress := buildIngressNginx(webInstall)
	if err := controllerutil.SetControllerReference(webInstall, ingress, r.Scheme); err != nil {
		return err
	}
	if err = r.Create(ctx, ingress); err != nil {
		r.Log.Error(err, "failed to create ingress-nginx resource")
		return err
	}
	r.Log.Info("ingress-nginx resource OK")
	return nil
}
func (r *WebInstallReconciler) updateDeploymentImage(ctx context.Context, deployment *apps.Deployment, webInstall *v1.WebInstall) error {
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		logrus.Warn("There is no image for the container(s) in the deployment resource")
		return nil
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != webInstall.Spec.Image {
		r.Log.WithField("old_image", deployment.Spec.Template.Spec.Containers[0].Image).Warn("new image detected --> changing")
		deployment = buildDeployment(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, deployment, r.Scheme); err != nil {
			return err
		}
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
	}

	return nil
}
func (r *WebInstallReconciler) updateService(ctx context.Context, webInstall *v1.WebInstall) error {
	service := &core.Service{}
	err := r.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, service)
	if err != nil && apierrors.IsNotFound(err) {
		r.Log.Warn("service not found for that RO --> creating one")
		service := buildService(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, service, r.Scheme); err != nil {
			return err
		}
		if err = r.Create(ctx, service); err != nil {
			r.Log.Error(err, "failed to create service resource at update")
			return err
		}
	} else if err != nil {
		return err
	}
	r.Log.Info("service resource OK")
	return nil
}
func (r *WebInstallReconciler) updateIngress(ctx context.Context, webInstall *v1.WebInstall) error {
	ingress := &net.Ingress{}
	err := r.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, ingress)
	if err != nil && apierrors.IsNotFound(err) {
		r.Log.Warn("ingress-nginx not found for that RO --> creating one")
		ingress := buildIngressNginx(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, ingress, r.Scheme); err != nil {
			return err
		}
		if err = r.Create(ctx, ingress); err != nil {
			r.Log.Error(err, "failed to create ingress resource")
			return err
		}
	} else if err != nil {
		return err
	}
	//Checking new Host updte...
	if len(ingress.Spec.Rules) == 0 {
		logrus.Warn("there is no rule for host in ingress resource")
		return nil
	}
	if ingress.Spec.Rules[0].Host != webInstall.Spec.Host {

		r.Log.WithField("old_host", ingress.Spec.Rules[0].Host).Warn("ingress-nginx host has been changed --> updating")
		ingress = buildIngressNginx(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, ingress, r.Scheme); err != nil {
			return err
		}
		if err := r.Update(ctx, ingress); err != nil {
			return err
		}

	}
	r.Log.Info("ingress resource OK")
	return nil
}
func (r *WebInstallReconciler) updateReplicas(ctx context.Context, deployment *apps.Deployment, webInstall *v1.WebInstall) error {
	if *deployment.Spec.Replicas != webInstall.Spec.Replicas {
		r.Log.WithField("old_count", *deployment.Spec.Replicas).Warn("incorrect replicas --> updating")

		deployment.Spec.Replicas = &webInstall.Spec.Replicas

		//Protection for deployment changes by kubernetes while running that reconcile procedure
		isUpdateSuccess := false
		for !isUpdateSuccess {
			err := r.Update(ctx, deployment)
			//Check error cause
			if err != nil && strings.Contains(err.Error(), "the object has been modified") {
				r.Log.Warn("deployment resource modified --> trying to Get the new version")
				err = r.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, deployment)
				if err != nil {
					r.Log.Error(err, "failed to get deployment resource")
					return err
				}
			} else if err != nil {
				r.Log.Error(err, "failed to update replica count")
				return err
			}
			isUpdateSuccess = true
		}
	}
	r.Log.Info("replica number OK")
	return nil

}
