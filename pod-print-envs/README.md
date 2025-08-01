# kubectl-pod-print-envs

`kubectl-pod-print-envs` is a `kubectl` plugin that allows you to interactively select a namespace and pod using fuzzy search (FZF-style), and then exec into the selected pod using `/bin/sh`.

This tool helps simplify navigating clusters with many namespaces and pods, especially when names are long or complex.

---

## ✨ Features

- Connects to the active kubeconfig context.
- Fuzzy-select from all available namespaces.
- Fuzzy-select from all running pods in a selected namespace.
- Automatically opens an interactive shell (`/bin/sh`) in the pod using `kubectl exec`.

---

## 🔧 Requirements

- Go 1.18+
- `kubectl` installed and configured.
- `~/.local/bin` is in your system `PATH` (for plugin discovery).
