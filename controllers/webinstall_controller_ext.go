package controllers

import (
	v1 "github.com/bartam1/kubopwebdep/api/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildDeployment(webInstall v1.WebInstall) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webInstall.Name + "-" + webInstall.Spec.Host,
			Namespace: webInstall.Namespace,
			//OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(&webInstall, v1.GroupVersion.WithKind("WebInstall"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &webInstall.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"webinstall.bartam/deployment-name": webInstall.Spec.Host,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"webinstall.bartam/deployment-name": webInstall.Spec.Host,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  webInstall.Spec.Host,
							Image: webInstall.Spec.Image,
						},
					},
				},
			},
		},
	}
	return &deployment
}

// cleanupOwnedResources will Delete any existing Deployment resources that
// were created for the given RO that no longer match the
// WebInstall.spec.host field.
/*
func (r *WebInstallReconciler) cleanupOwnedResources(ctx context.Context, log logr.Logger, webInstall *v1.WebInstall) error {
	log.Info("finding old Deployments for WebInstall resource...")

	//Get deployment resource owned by this RO
	var deployments apps.DeploymentList
	if err := r.Client.List(ctx, &deployments, client.InNamespace(webInstall.Namespace)); err != nil {
		return err
	}
	log.Info("dep size", "s", len(deployments.Items))
	deleted := 0
	for _, depl := range deployments.Items {
		//log.Info("dep size", deployments.Size())
		if depl.Name == webInstall.Spec.Host {
			// If this deployment's name matches the one on the WebInstall resource
			// then do not delete it.
			continue
		}

		if err := r.Client.Delete(ctx, &depl); err != nil {
			log.Error(err, "failed to delete Deployment resource")
			return err
		}
		deleted++
	}
	if deleted > 0 {
		log.Info("finished cleaning up old Deployment resources", "number_deleted", deleted)
	} else {
		log.Info("not found any old deployment resources for WebInstall")
	}

	return nil
}
*/
