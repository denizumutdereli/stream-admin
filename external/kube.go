package main

/*
TODO: refactoring as service layer and DI with relavant registry and builders to obtain specific tasks.
*/

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// TODO: kubeconfig
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		log.Fatal("Cannot find home directory")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Could not create Kubernetes client: %s", err)
	}

	// TODO: permission logic
	fmt.Println("Building Docker image...")
	cmd := exec.Command("docker", "build", "-t", "python-app", ".")
	cmd.Dir = "./external/python/app"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not build Docker image: %s", err)
	}

	deploymentClient := clientset.AppsV1().Deployments(corev1.NamespaceDefault)
	serviceClient := clientset.CoreV1().Services(corev1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "python-app-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "python-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "python-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "python-app",
							Image: "python-app:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 9090,
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "python-app-service",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
					NodePort:   30090,
				},
			},
			Selector: map[string]string{
				"app": "python-app",
			},
		},
	}

	fmt.Println("Creating deployment...")
	result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Could not create deployment: %s", err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	fmt.Println("Creating service...")
	resultService, err := serviceClient.Create(context.TODO(), service, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Could not create service: %s", err)
	}
	fmt.Printf("Created service %q.\n", resultService.GetObjectMeta().GetName())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Cleanup
	fmt.Println("\nDeleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentClient.Delete(context.TODO(), "python-app-deployment", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		log.Fatalf("Could not delete deployment: %s", err)
	}
	fmt.Println("Deleting service...")
	if err := serviceClient.Delete(context.TODO(), "python-app-service", metav1.DeleteOptions{}); err != nil {
		log.Fatalf("Could not delete service: %s", err)
	}
	fmt.Println("Deleted deployment and service.")
}

func int32Ptr(i int32) *int32 { return &i }
