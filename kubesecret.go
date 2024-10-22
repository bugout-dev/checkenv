package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubernetesSecretProvider(filter string) (map[string]string, error) {
	var kubeconfig string
	var namespace string
	var secret string

	vars := map[string]string{}

	filterStrings := strings.Split(filter, "+")
	for _, filterString := range filterStrings {
		components := strings.Split(filterString, ":")
		k := components[0]
		v := strings.Join(components[1:], ":")
		switch k {
		case "kubeconfig":
			// TODO(zomglings): "+" is a valid symbol in file/directory names. The current way of separating args using "+" is not robust.
			kubeconfig = v
		case "namespace":
			namespace = v
		case "secret":
			secret = v
		}
	}

	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			defaultConfig := filepath.Join(home, ".kube", "config")
			defaultConfigStat, defaultConfigErr := os.Stat(defaultConfig)
			if defaultConfigErr != nil {
				return vars, fmt.Errorf("kubeconfig not provided and default kubeconfig (%s) not found", defaultConfig)
			}
			if defaultConfigStat.IsDir() {
				return vars, fmt.Errorf("kubeconfig not provided and default kubeconfig (%s) is a directory", defaultConfig)
			}
		}
	}

	if kubeconfig == "" {
		return vars, errors.New("kubeconfig not provided")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secretsClient := clientset.CoreV1().Secrets(namespace)

	ctx := context.Background()

	if secret != "" {
		secretResource, secretResourceErr := secretsClient.Get(ctx, secret, metav1.GetOptions{})
		if secretResourceErr != nil {
			return vars, secretResourceErr
		}
		for key, value := range secretResource.Data {
			vars[key] = string(value)
		}
	} else {
		secretResources, secretResourcesErr := secretsClient.List(ctx, metav1.ListOptions{})
		if secretResourcesErr != nil {
			return vars, secretResourcesErr
		}
		for _, item := range secretResources.Items {
			for key, value := range item.Data {
				vars[key] = string(value)
			}
		}
	}

	return vars, nil
}

func init() {
	helpString := "Provides variables defined in secrets on a Kubernetes cluster, filterable by namespace and by secret name."
	RegisterPlugin("kubesecret", helpString, KubernetesSecretProvider)
}
