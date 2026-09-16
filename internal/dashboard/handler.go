package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/preview"
)

// Handler serves the dashboard JSON API and static UI from a Kubernetes client.
// All of its operations are read-only.
type Handler struct {
	Clientset kubernetes.Interface
	Logger    *slog.Logger
	TTL       time.Duration
}

// data writes the current set of preview environments as JSON. It is a pure
// read: it lists labeled namespaces and reports each environment's status.
func (h *Handler) data(w http.ResponseWriter, r *http.Request) {
	resp, err := h.collect(r.Context())
	if err != nil {
		h.Logger.Error("dashboard collect", "error", err)
		http.Error(w, "cannot load environments", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.Logger.Error("dashboard encode", "error", err)
	}
}

// collect lists preview namespaces and builds the dashboard payload.
func (h *Handler) collect(ctx context.Context) (APIResponse, error) {
	list, err := h.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: preview.LabelPR,
	})
	if err != nil {
		return APIResponse{}, err
	}
	now := time.Now()
	envs := make([]EnvView, 0, len(list.Items))
	for i := range list.Items {
		ns := &list.Items[i]
		pr, err := strconv.Atoi(ns.Labels[preview.LabelPR])
		if err != nil {
			continue
		}
		svc := ns.Labels[preview.LabelService]
		host := fmt.Sprintf("pr-%d-%s.127-0-0-1.nip.io", pr, svc)
		view := EnvView{
			PR:        pr,
			Service:   svc,
			Namespace: ns.Name,
			URL:       "http://" + host,
			Status:    h.status(ctx, ns.Name, svc),
		}
		if created := createdAt(ns); !created.IsZero() {
			view.CreatedAt = created
			view.AgeSeconds = int64(now.Sub(created).Seconds())
			if remaining := h.TTL - now.Sub(created); remaining > 0 {
				view.ExpiresInSeconds = int64(remaining.Seconds())
			}
		}
		envs = append(envs, view)
	}
	sort.Slice(envs, func(i, j int) bool { return envs[i].PR < envs[j].PR })
	return APIResponse{
		Summary:      Summary{Active: len(envs), GeneratedAt: now},
		Environments: envs,
	}, nil
}

// status reports a human-friendly state for a preview from its Deployment. A
// missing Deployment is reported as "Provisioning".
func (h *Handler) status(ctx context.Context, namespace, svc string) string {
	dep, err := h.Clientset.AppsV1().Deployments(namespace).Get(ctx, svc, metav1.GetOptions{})
	if err != nil {
		return "Provisioning"
	}
	for _, c := range dep.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
			return "Live"
		}
	}
	return "Deploying"
}

// createdAt parses the Unix-seconds creation label; a zero time means unknown.
func createdAt(ns *corev1.Namespace) time.Time {
	raw, ok := ns.Labels[preview.LabelCreatedAt]
	if !ok {
		return time.Time{}
	}
	secs, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}