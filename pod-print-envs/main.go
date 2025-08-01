package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var grep string
	var igrep string
	var rootCmd = &cobra.Command{
		Use:   "pod-print-env",
		Short: "Print environment variables from a pod",
		Run: func(cmd *cobra.Command, args []string) {
			if grep != "" {
				fmt.Printf("Filtering with grep: '%s'\n", grep)
			}
			if igrep != "" {
				fmt.Printf("Filtering with igrep: '%s'\n", igrep)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&grep, "grep", "", "Filter output by pattern")
	rootCmd.PersistentFlags().StringVar(&igrep, "igrep", "", "Filter output by pattern case insensitive")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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

	// Execute env in pod
	command := "env"
	if grep != "" {
		command = "env | grep " + grep
	}

	if igrep != "" {
		command = "env | grep -i " + igrep
	}
	cmd := exec.Command("kubectl", "exec", "-n", namespace, pod, "--", "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Kubectl Command: %s\n", cmd)
	fmt.Printf("Environment variables in the pod %s/%s:\n", namespace, pod)
	fmt.Println("-------------------------------------------")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to exec into pod: %v", err)
	}
}
