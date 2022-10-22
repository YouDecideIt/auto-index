package operations

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap/tidb-operator/pkg/apis/pingcap/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	scheme = runtime.NewScheme()
)

func restoreTime() string {
	return time.Now().Add(-3 * time.Minute).Format("2006-01-02 15:04:05-0700")
}

func New() BCluster {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.SchemeBuilder.AddToScheme(scheme))
	c, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	return &tcOp{
		c: c,
	}
}

type tcOp struct {
	c client.Client
}

func newNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mirror",
		},
	}
}

func newServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "mirror",
			Annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "arn:aws:iam::288114636544:role/hackathon-tidb-mirror",
			},
		},
	}
}

func (op *tcOp) DestroyBCluster(ctx context.Context) error {
	ns := newNamespace()
	if err := op.c.Delete(ctx, ns); err != nil {
		return err
	}

	return nil
}

func (op *tcOp) StartBCluster(ctx context.Context) error {
	fmt.Println("create ns")
	ns := newNamespace()
	if err := op.c.Create(ctx, ns); err != nil {
		return err
	}

	fmt.Println("create sa")
	sa := newServiceAccount()
	if err := op.c.Update(ctx, sa); err != nil {
		return err
	}

	fmt.Println("create tc")
	tc := newTC()
	if err := op.c.Create(ctx, tc); err != nil {
		return err
	}
	fmt.Println("create tc svc")
	svc := newPublicService()
	if err := op.c.Create(ctx, svc); err != nil {
		return err
	}
	fmt.Println("create ng")
	ngMonitor := newNGMonitor()
	if err := op.c.Create(ctx, ngMonitor); err != nil {
		return err
	}

	fmt.Println("create ng svc")
	ngSvc := newNGMonitorPublicService()
	if err := op.c.Create(ctx, ngSvc); err != nil {
		return err
	}

	fmt.Println("create dashboard svc")
	dashboardSvc := newDashboardPublicService()
	if err := op.c.Create(ctx, dashboardSvc); err != nil {
		return err
	}

	return nil
}

func (op *tcOp) restoreToBCluster(ctx context.Context) error {
	job := newRestoreJob(restoreTime())
	if err := op.c.Create(ctx, job); err != nil {
		return err
	}

	if err := WaitForCondition(ctx, op.c, job, func() bool {
		for _, cond := range job.Status.Conditions {
			if cond.Type == batchv1.JobComplete &&
				cond.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	}, DefaultTimeout); err != nil {
		return err
	}

	return nil
}

func (op *tcOp) WaitBClusterStartedAndMirrored(ctx context.Context) (*Cluster, error) {
	c := &Cluster{}
	fmt.Println("wait b cluster ready")
	tc := &v1alpha1.TidbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror",
			Namespace: "mirror",
		},
	}
	if err := WaitForCondition(ctx, op.c, tc, func() bool {
		for _, cond := range tc.Status.Conditions {
			if cond.Type == v1alpha1.TidbClusterReady &&
				cond.Status == corev1.ConditionTrue {
				return true

			}
		}

		return false
	}, DefaultTimeout); err != nil {
		return nil, err
	}

	fmt.Println("restore to b cluster")
	if err := op.restoreToBCluster(ctx); err != nil {
		return nil, err
	}

	fmt.Println("wait tidb service ready")
	tcSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror-tidb-public",
			Namespace: "mirror",
		},
	}
	if err := WaitForCondition(ctx, op.c, tcSvc, func() bool {
		if len(tcSvc.Status.LoadBalancer.Ingress) == 0 {
			return false
		}

		ingress := tcSvc.Status.LoadBalancer.Ingress[0]
		address := ingress.Hostname
		if address == "" {
			return false
		}

		dsn := fmt.Sprintf("root@tcp(%s:4000)/?timeout=5s", address)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return false
		}
		db.Close()

		c.SQLEndpoint = address

		return true
	}, DefaultTimeout); err != nil {
		return nil, err
	}

	fmt.Println("wait ng monitoring service ready")
	ngSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror-ng-monitoring-public",
			Namespace: "mirror",
		},
	}

	if err := WaitForCondition(ctx, op.c, ngSvc, func() bool {
		if len(ngSvc.Status.LoadBalancer.Ingress) == 0 {
			return false
		}

		ingress := ngSvc.Status.LoadBalancer.Ingress[0]
		address := ingress.Hostname
		if address == "" {
			return false
		}

		c.NgmEndpoint = address

		return true
	}, DefaultTimeout); err != nil {
		return nil, err
	}

	return c, nil
}

func (op *tcOp) CreateTiDBCluster(ctx context.Context, c *Cluster) (*Cluster, error) {
	tc := newTC()
	if err := op.c.Create(ctx, tc); err != nil {
		return nil, err
	}
	svc := newPublicService()
	if err := op.c.Create(ctx, svc); err != nil {
		return nil, err
	}

	if err := WaitForCondition(ctx, op.c, tc, func() bool {
		for _, cond := range tc.Status.Conditions {
			if cond.Type == v1alpha1.TidbClusterReady &&
				cond.Status == corev1.ConditionTrue {
				return true

			}
		}

		return false
	}, DefaultTimeout); err != nil {
		return nil, err
	}

	if err := WaitForCondition(ctx, op.c, svc, func() bool {
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return false
		}

		ingress := svc.Status.LoadBalancer.Ingress[0]
		address := ingress.Hostname
		if address == "" {
			return false
		}

		dsn := fmt.Sprintf("root@tcp(%s:4000)/?timeout=5s", address)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return false
		}
		db.Close()

		return true
	}, DefaultTimeout); err != nil {
		return nil, err
	}

	return c, nil
}

func mustNewTiKVConfig(data []byte) *v1alpha1.TiKVConfigWraper {
	cfg := v1alpha1.NewTiKVConfig()
	if err := cfg.UnmarshalTOML(data); err != nil {
		panic(err)
	}
	return cfg
}

const (
	defaultTiKVConfig = `
[storage]
reserve-space = "0MB"

[rocksdb]
max-open-files = 256

[raftdb]
max-open-files = 256
    `
)

func newTC() *v1alpha1.TidbCluster {
	policy := corev1.PersistentVolumeReclaimDelete
	return &v1alpha1.TidbCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror",
			Namespace: "mirror",
		},
		Spec: v1alpha1.TidbClusterSpec{
			Version:                    "v6.3.0",
			PVReclaimPolicy:            &policy,
			EnableDynamicConfiguration: pointer.Bool(true),
			Helper: &v1alpha1.HelperSpec{
				Image: pointer.String("alpine:3.16.0"),
			},
			PD: &v1alpha1.PDSpec{
				EnableDashboardInternalProxy: pointer.Bool(true),
				BaseImage:                    "pingcap/pd",
				MaxFailoverCount:             pointer.Int32(0),
				Replicas:                     1,
				ResourceRequirements: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("100Gi"),
					},
				},
				Config: v1alpha1.NewPDConfig(),
			},
			TiKV: &v1alpha1.TiKVSpec{
				BaseImage:          "pingcap/tikv",
				MaxFailoverCount:   pointer.Int32(0),
				EvictLeaderTimeout: pointer.String("1m"),
				Replicas:           3,
				ResourceRequirements: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("100Gi"),
						corev1.ResourceCPU:     resource.MustParse("3"),
					},
				},
				Config: mustNewTiKVConfig([]byte(defaultTiKVConfig)),
			},
			TiDB: &v1alpha1.TiDBSpec{
				BaseImage:        "pingcap/tidb",
				MaxFailoverCount: pointer.Int32(0),
				Replicas:         2,
				Service: &v1alpha1.TiDBServiceSpec{
					ServiceSpec: v1alpha1.ServiceSpec{
						Type: corev1.ServiceTypeClusterIP,
					},
				},
				Config: v1alpha1.NewTiDBConfig(),
			},
		},
	}
}

func newPublicService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror-tidb-public",
			Namespace: "mirror",
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type":                              "external",
				"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type":                   "instance",
				"service.beta.kubernetes.io/aws-load-balancer-scheme":                            "internet-facing",
				"service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled": "true",
				"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes":           "preserve_client_ip.enabled=true",
			},
		},
		Spec: corev1.ServiceSpec{
			LoadBalancerSourceRanges: []string{"0.0.0.0/0"},
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyTypeLocal,
			Ports: []corev1.ServicePort{
				{
					Name:       "mysql-client",
					Port:       4000,
					TargetPort: intstr.FromInt(4000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/component": "tidb",
				"app.kubernetes.io/instance":  "mirror",
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}
}

func newNGMonitor() *v1alpha1.TidbNGMonitoring {
	return &v1alpha1.TidbNGMonitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror",
			Namespace: "mirror",
		},
		Spec: v1alpha1.TidbNGMonitoringSpec{
			Clusters: []v1alpha1.TidbClusterRef{
				{
					Name: "mirror",
				},
			},
			NGMonitoring: v1alpha1.NGMonitoringSpec{
				ResourceRequirements: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("100Gi"),
					},
				},
				ComponentSpec: v1alpha1.ComponentSpec{
					Version: pointer.String("v6.3.0"),
				},
				BaseImage: "pingcap/ng-monitoring",
			},
		},
	}
}

func newNGMonitorPublicService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror-ng-monitoring-public",
			Namespace: "mirror",
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type":                              "external",
				"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type":                   "instance",
				"service.beta.kubernetes.io/aws-load-balancer-scheme":                            "internet-facing",
				"service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled": "true",
				"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes":           "preserve_client_ip.enabled=true",
			},
		},
		Spec: corev1.ServiceSpec{
			LoadBalancerSourceRanges: []string{"0.0.0.0/0"},
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyTypeLocal,
			Ports: []corev1.ServicePort{
				{
					Name:       "ng-monitoring",
					Port:       12020,
					TargetPort: intstr.FromInt(12020),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/component": "ng-monitoring",
				"app.kubernetes.io/instance":  "mirror",
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

}

func newRestoreJob(ts string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "restore",
			Namespace: "mirror",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: pointer.Int32(1),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "restore",
							Image: "pingcap/br:v6.3.0",
							Command: []string{
								"/br",
								"restore",
								"point",
								"--pd",
								"mirror-pd:2379",
								"--full-backup-storage",
								"s3://hackathon-youdecideit/full-test",
								"--storage",
								"s3://hackathon-youdecideit/log",
								"--send-credentials-to-tikv=false",
								"--restored-ts",
								ts,
							},
						},
					},
				},
			},
		},
	}
}

func newDashboardPublicService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mirror-tidb-dashboard-public",
			Namespace: "mirror",
			Annotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-type":                              "external",
				"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type":                   "instance",
				"service.beta.kubernetes.io/aws-load-balancer-scheme":                            "internet-facing",
				"service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled": "true",
				"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes":           "preserve_client_ip.enabled=true",
			},
		},
		Spec: corev1.ServiceSpec{
			LoadBalancerSourceRanges: []string{"0.0.0.0/0"},
			ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyTypeLocal,
			Ports: []corev1.ServicePort{
				{
					Name:       "dashboard",
					Port:       10262,
					TargetPort: intstr.FromInt(10262),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/component": "discovery",
				"app.kubernetes.io/instance":  "mirror",
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

}
