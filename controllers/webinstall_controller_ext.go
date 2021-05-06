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

func (r *WebInstallReconciler) createWebInstallBundle(ctx context.Context, webInstall *v1.WebInstall, logger *logrus.Entry) (err error) {
	deployment := buildDeployment(webInstall)
	//Bind webInstall to that deployment for reconcile at changes
	if err := controllerutil.SetControllerReference(webInstall, deployment, r.Scheme); err != nil {
		// requeue with error
		return err
	}
	logger.Info("deployment OK")
	if err := r.Client.Create(ctx, deployment); err != nil {
		logger.Error(err, "failed to create Deployment resource")
		return err
	}
	service := buildService(webInstall)
	if err := controllerutil.SetControllerReference(webInstall, service, r.Scheme); err != nil {
		return err
	}
	if err = r.Create(ctx, service); err != nil {
		logger.Error(err, "failed to create service for Deployment")
		return err
	}
	logger.Info("service OK")
	ingress := buildIngressNginx(webInstall)
	if err := controllerutil.SetControllerReference(webInstall, ingress, r.Scheme); err != nil {
		return err
	}
	if err = r.Create(ctx, ingress); err != nil {
		logger.Error(err, "failed to create ingress-nginx for Deployment")
		return err
	}
	logger.Info("ingress-nginx OK")
	return nil
}
func (r *WebInstallReconciler) updateDeploymentImage(ctx context.Context, deployment *apps.Deployment, webInstall *v1.WebInstall, logger *logrus.Entry) error {
	if deployment.Spec.Template.Spec.Containers[0].Image != webInstall.Spec.Image {
		logger.Warn("new image detected --> changing")
		deployment = buildDeployment(webInstall)
		if err := r.Client.Update(ctx, deployment); err != nil {
			return err
		}
	}
	return nil
}
func (r *WebInstallReconciler) updateService(ctx context.Context, webInstall *v1.WebInstall, logger *logrus.Entry) error {
	service := &core.Service{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, service)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Warn("service not found for that RO --> creating one")
		service := buildService(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, service, r.Scheme); err != nil {
			return err
		}
		if err = r.Client.Create(ctx, service); err != nil {
			logger.Error(err, "failed to create service for Deployment")
			return err
		}
	} else if err != nil {
		return err
	}
	logger.Info("service resource OK")
	return nil
}
func (r *WebInstallReconciler) updateIngress(ctx context.Context, webInstall *v1.WebInstall, logger *logrus.Entry) error {
	ingress := &net.Ingress{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, ingress)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Warn("ingress-nginx not found for that RO --> creating one")
		ingress := buildIngressNginx(webInstall)
		if err := controllerutil.SetControllerReference(webInstall, ingress, r.Scheme); err != nil {
			return err
		}
		if err = r.Client.Create(ctx, ingress); err != nil {
			logger.Error(err, "failed to create service for Deployment")
			return err
		}
	} else if err != nil {
		return err
	}
	//Checking new Host updte...
	if len(ingress.Spec.Rules) > 0 && ingress.Spec.Rules[0].Host != webInstall.Spec.Host {

		logger.Warn("ingress-nginx host has changed --> updating")
		ingress = buildIngressNginx(webInstall)
		if err := r.Client.Update(ctx, ingress); err != nil {
			return err
		}

	}
	logger.Info("ingress resource OK")
	return nil
}
func (r *WebInstallReconciler) updateReplicas(ctx context.Context, deployment *apps.Deployment, webInstall *v1.WebInstall, logger *logrus.Entry) error {
	if *deployment.Spec.Replicas != webInstall.Spec.Replicas {
		logger.Warn("incorrect replicas --> updating", "old_count", *deployment.Spec.Replicas, "new_count", webInstall.Spec.Replicas)

		deployment.Spec.Replicas = &webInstall.Spec.Replicas

		//Protection for deployment changes by kubernetes while running that reconcile
		isUpdateSuccess := false
		for !isUpdateSuccess {
			err := r.Client.Update(ctx, deployment)
			if err != nil && strings.Contains(err.Error(), "the object has been modified") {
				logger.Warn("deployment resource modified --> trying to Get the new version")
				err = r.Client.Get(ctx, client.ObjectKey{Namespace: webInstall.Namespace, Name: webInstall.Name}, deployment)
				if err != nil {
					logger.Error(err, "failed to get deployment")
					return err
				}
			} else if err != nil {
				logger.Error(err, "failed to update replica count")
				return err
			}
			isUpdateSuccess = true
		}
	}
	logger.Info("replica number OK")
	return nil

}
