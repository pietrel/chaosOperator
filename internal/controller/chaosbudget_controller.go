package controller

import (
	"context"
	"time"

	v1 "chaosOperator/api/v1"
	"chaosOperator/internal/budget"
	"chaosOperator/internal/metrics"
	"chaosOperator/internal/policy"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ChaosBudgetReconciler reconciles a ChaosBudget object
type ChaosBudgetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Metrics  metrics.Client
	Budget   budget.Calculator
	Policy   policy.Evaluator
	Interval time.Duration
}

// +kubebuilder:rbac:groups=chaos.example.com,resources=chaosbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=chaos.example.com,resources=chaosbudgets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=chaos.example.com,resources=chaosbudgets/finalizers,verbs=update

func (r *ChaosBudgetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	var cb v1.ChaosBudget
	if err := r.Get(ctx, req.NamespacedName, &cb); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling ChaosBudget", "name", cb.Name)

	// Fetch metrics
	m, err := r.Metrics.FetchMetric(ctx, cb.Spec.Budget.Type, cb.Spec.Window, cb.Spec.Target.Namespace, cb.Spec.Target.Labels)
	if err != nil {
		l.Error(err, "Failed to fetch metrics, defaulting to denied")
		cb.Status.Allowed = false
		cb.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, &cb); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: r.Interval}, nil
	}

	// Calculate usage
	consumed := r.Budget.Calculate(cb.Spec.Budget, m)
	remaining := cb.Spec.Budget.Max - consumed

	// Policy evaluation
	policyTypes := make([]string, len(cb.Spec.Policies))
	for i, p := range cb.Spec.Policies {
		policyTypes[i] = p.Type
	}
	decision := r.Policy.Evaluate(policyTypes, remaining)

	// Update status
	cb.Status.Consumed = consumed
	cb.Status.Remaining = remaining
	cb.Status.Allowed = decision.Allowed
	cb.Status.LastUpdated = metav1.Now()

	l.Info("Updating status", "consumed", consumed, "remaining", remaining, "allowed", decision.Allowed, "reason", decision.Reason)

	if err := r.Status().Update(ctx, &cb); err != nil {
		l.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.Interval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChaosBudgetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ChaosBudget{}).
		Complete(r)
}
