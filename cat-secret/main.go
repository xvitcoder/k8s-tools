package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func printHelp() {
	fmt.Println(`Usage:
  kubectl cat-secret

This plugin lets you fuzzy-select a namespace and secret, then print it in the sysout.`)
}

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "help" || arg == "-h" {
			printHelp()
			return
		}
	}

	// Load kubeconfig from default location
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Fatalf("Error loading kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Fetch namespaces
	nsList, err := clientset.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Error fetching namespaces: %v", err)
	}
	namespaces := make([]string, len(nsList.Items))
	for i, ns := range nsList.Items {
		namespaces[i] = ns.Name
	}

	// Fuzzy-select namespace
	nsIdx, err := fuzzyfinder.Find(
		namespaces,
		func(i int) string { return namespaces[i] },
		fuzzyfinder.WithPromptString("Select namespace > "),
	)
	if err != nil {
		log.Fatalf("Namespace selection aborted: %v", err)
	}
	namespace := namespaces[nsIdx]

	// List secrets
	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing secrets: %v", err)
	}
	if len(secrets.Items) == 0 {
		log.Fatalf("No secrets in namespace %s", namespace)
	}

	secretIndex, err := fuzzyfinder.Find(
		secrets.Items,
		func(i int) string {
			return secrets.Items[i].Name
		},
		fuzzyfinder.WithPromptString("Select secret > "),
	)
	if err != nil {
		log.Fatalf("No secret selected")
	}
	secret := secrets.Items[secretIndex]

	fmt.Printf("Contents of secret %s in namespace %s:\n", secret.Name, namespace)
	fmt.Println("-------------------------------------------")

	for key, val := range secret.Data {
		fmt.Printf("%s: %s\n", key, string(val))
	}
}
