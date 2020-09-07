package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.xrstf.de/kuwait/pkg/condition"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type options struct {
	kubeconfig string
	timeout    time.Duration
	interval   time.Duration
	debugLog   bool
}

func main() {
	opt := options{
		timeout:  10 * time.Minute,
		interval: 1 * time.Second,
	}

	flag.StringVar(&opt.kubeconfig, "kubeconfig", opt.kubeconfig, "kubeconfig file to use")
	flag.DurationVar(&opt.timeout, "timeout", opt.timeout, "maximum time to wait for all conditions to be met")
	flag.DurationVar(&opt.interval, "interval", opt.interval, "time inbetween status checks")
	flag.BoolVar(&opt.debugLog, "debug", opt.debugLog, "enable more verbose logging")
	flag.Parse()

	// setup logging
	var log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC1123,
	})

	if opt.debugLog {
		log.SetLevel(logrus.DebugLevel)
	}

	// validate CLI flags
	conditionArgs := flag.Args()
	if len(conditionArgs) == 0 {
		log.Fatal("No conditions to wait for given.")
	}

	// setup signal handler
	stopCh := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-stopCh
		cancel()
	}()

	// setup kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", opt.kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes clientset: %v", err)
	}

	log.Debug("Creating REST mapper...")

	mapper, err := getRESTMapper(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes REST mapper: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed dynamic.NewForConfig: %v", err)
	}

	log.Debug("Parsing conditions...")

	conditions := []condition.Condition{}
	for _, conditionArg := range conditionArgs {
		cond, err := condition.Parse(conditionArg, clientset, dynamicClient, mapper)
		if err != nil {
			log.Fatalf("Invalid condition %q: %v", conditionArg, err)
		}

		log.Debugf("Condition: %v", cond)

		conditions = append(conditions, cond)
	}

	// start timeout
	to, tocancel := context.WithTimeout(ctx, 5*time.Second)
	defer tocancel()

	log.Infof("Waiting %v for the following conditions to be met:", opt.timeout)
	for _, condition := range conditions {
		log.Info(condition.String())
	}

	// start goroutine per condition
	wg := sync.WaitGroup{}
	success := true
	for i := range conditions {
		wg.Add(1)
		go func(condition condition.Condition) {
			if err := waiter(to, log, condition, opt.interval); err != nil {
				log.Error(err)
				success = false
			}

			wg.Done()
		}(conditions[i])
	}

	wg.Wait()

	if success {
		log.Info("All conditions are met.")
	} else {
		os.Exit(1)
	}
}

func waiter(ctx context.Context, log logrus.FieldLogger, condition condition.Condition, interval time.Duration) error {
	err := wait.PollUntil(interval, func() (bool, error) {
		return condition.Satisfied(ctx)
	}, ctx.Done())

	if err == wait.ErrWaitTimeout || err == context.DeadlineExceeded {
		return fmt.Errorf("Condition was not met after timeout: %s", condition)
	}

	return err
}

func getRESTMapper(config *rest.Config) (meta.RESTMapper, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	cache := memory.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cache)
	fancyMapper := restmapper.NewShortcutExpander(mapper, discoveryClient)

	return fancyMapper, nil
}
