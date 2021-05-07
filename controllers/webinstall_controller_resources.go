package controllers

import (
	v1 "github.com/bartam1/kubopwebdep/api/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	net "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func buildService(webInstall *v1.WebInstall) *core.Service {
	service := core.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webInstall.Name,
			Namespace: webInstall.Namespace,
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				core.ServicePort{
					Port: v1.HostPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Type(0),
						IntVal: v1.ContainerPort,
					},
				},
			},
			Selector: map[string]string{
				"webinstall.bartam/deployment-name": webInstall.Name,
			},
			HealthCheckNodePort: 0,
		},
	}

	return &service
}
func buildIngressNginx(webInstall *v1.WebInstall) *net.Ingress {
	ingress := &net.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webInstall.Name,
			Namespace: webInstall.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":    "nginx",
				"cert-manager.io/cluster-issuer": "letsencrypt-prod",
			},
		},
		Spec: net.IngressSpec{
			TLS: []net.IngressTLS{
				net.IngressTLS{
					Hosts: []string{
						webInstall.Spec.Host,
					},
					SecretName: webInstall.Spec.Host + "-webinst-secret",
				},
			},

			Rules: []net.IngressRule{
				net.IngressRule{
					Host: webInstall.Spec.Host,
					IngressRuleValue: net.IngressRuleValue{
						HTTP: &net.HTTPIngressRuleValue{
							Paths: []net.HTTPIngressPath{
								net.HTTPIngressPath{
									Path: "/",
									Backend: net.IngressBackend{
										ServiceName: webInstall.Name,
										ServicePort: intstr.IntOrString{
											Type:   intstr.Type(0),
											IntVal: v1.HostPort,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}
func buildDeployment(webInstall *v1.WebInstall) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webInstall.Name,
			Namespace: webInstall.Namespace,
		},
		Spec: apps.DeploymentSpec{
			Replicas: &webInstall.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"webinstall.bartam/deployment-name": webInstall.Name,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"webinstall.bartam/deployment-name": webInstall.Name,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  webInstall.Name,
							Image: webInstall.Spec.Image,
							Ports: []core.ContainerPort{{ContainerPort: v1.ContainerPort}},
						},
					},
				},
			},
		},
	}
	return &deployment
}
