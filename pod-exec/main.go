package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func printHelp() {
	fmt.Println(`Usage:
  kubectl pod-exec

This plugin lets you fuzzy-select a namespace and pod, then exec into the pod.`)
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

	// Fetch pods in selected namespace
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatalf("Error fetching pods: %v", err)
	}
	if len(podList.Items) == 0 {
		fmt.Println("No pods found in namespace:", namespace)
		os.Exit(0)
	}

	pods := make([]string, len(podList.Items))
	for i, pod := range podList.Items {
		pods[i] = pod.Name
	}

	// Fuzzy-select pod
	podIdx, err := fuzzyfinder.Find(
		pods,
		func(i int) string { return pods[i] },
		fuzzyfinder.WithPromptString("Select pod > "),
	)
	if err != nil {
		log.Fatalf("Pod selection aborted: %v", err)
	}
	pod := pods[podIdx]

	// Execute shell in pod
	shell := "/bin/sh"
	cmd := exec.Command("kubectl", "exec", "-n", namespace, "-it", pod, "--", shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Executing shell in pod %s/%s...\n", namespace, pod)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to exec into pod: %v", err)
	}
}
