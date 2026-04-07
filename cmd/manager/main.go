package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	v1 "chaosOperator/api/v1"
	"chaosOperator/internal/api"
	"chaosOperator/internal/budget"
	"chaosOperator/internal/controller"
	"chaosOperator/internal/metrics"
	"chaosOperator/internal/policy"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var promAddr string
	var apiAddr string
	var reconcileInterval time.Duration

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&promAddr, "prometheus-address", "http://prometheus-k8s.monitoring.svc:9090", "Prometheus server address")
	flag.StringVar(&apiAddr, "api-bind-address", ":8082", "The address the decision API binds to.")
	flag.DurationVar(&reconcileInterval, "reconcile-interval", 60*time.Second, "Reconcile interval")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "chaos-budget-operator.example.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	promClient, err := metrics.NewClient(promAddr)
	if err != nil {
		setupLog.Error(err, "unable to create prometheus client")
		os.Exit(1)
	}

	if err = (&controller.ChaosBudgetReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Metrics:  promClient,
		Budget:   budget.NewCalculator(),
		Policy:   policy.NewEvaluator(),
		Interval: reconcileInterval,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ChaosBudget")
		os.Exit(1)
	}

	// Start Agent Communication API
	apiServer := &api.Server{Client: mgr.GetClient()}
	http.HandleFunc("/check", apiServer.Check)
	go func() {
		setupLog.Info("starting API server", "addr", apiAddr)
		if err := http.ListenAndServe(apiAddr, nil); err != nil {
			setupLog.Error(err, "API server failed")
		}
	}()

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
