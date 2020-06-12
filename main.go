package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

type options struct {
	kubeconfig string
	timeout    time.Duration
	debugLog   bool
}

func main() {
	opt := options{
		timeout: 10 * time.Minute,
	}

	flag.StringVar(&opt.kubeconfig, "kubeconfig", opt.kubeconfig, "kubeconfig file to use")
	flag.DurationVar(&opt.timeout, "timeout", opt.timeout, "maximum time to wait for all conditions to be met")
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
	targetArgs := flag.Args()
	if len(targetArgs) == 0 {
		log.Fatal("No targets to wait for given.")
	}

	targets := []target{}
	for _, target := range targetArgs {
		t, err := parseTarget(target)
		if err != nil {
			log.Fatalf("Invalid target %q: %v", target, err)
		}

		targets = append(targets, t)
	}

	// setup kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", opt.kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	fmt.Println(config.String())
}

type target struct {
	kind      string
	namespace string
	name      string
	condition string
}

const (
	KindNamespace   = "namespace"
	KindPod         = "pod"
	KindDaemonSet   = "daemonset"
	KindStatefulSet = "statefulset"
	KindDeployment  = "deployment"
)

func parseTarget(t string) (target, error) {
	result := target{}

	parts := strings.Split(t, "/")
	if len(parts) < 4 {
		return result, errors.New("must be in `kind/namespace/name/condition` format")
	}

	result.kind = strings.ToLower(parts[0])
	result.namespace = parts[1]
	result.name = parts[2]
	result.condition = parts[3]

	switch result.kind {
	case "pod":
		fallthrough
	case "pods":
		result.kind = KindPod

	case "namespace":
		fallthrough
	case "namespaces":
		fallthrough
	case "ns":
		result.kind = KindNamespace

	case "daemonset":
		fallthrough
	case "daemonsets":
		fallthrough
	case "ds":
		result.kind = KindDaemonSet

	case "statefulset":
		fallthrough
	case "statefulsets":
		fallthrough
	case "sts":
		fallthrough
	case "ss":
		result.kind = KindStatefulSet

	case "deployment":
		fallthrough
	case "deployments":
		fallthrough
	case "dep":
		fallthrough
	case "deps":
		result.kind = KindDeployment

	default:
		return result, fmt.Errorf("unknown kind %q given", result.kind)
	}

	return result, nil
}
