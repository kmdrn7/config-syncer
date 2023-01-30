package main

import (
	"fmt"
	"os"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	flag "github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"
)

// TODO: wrap in cobra style
// TODO: implement sync using configurable YAML config file instead of using src-secret and dest-secret
// TODO: prepare helm charts

var (
	client     *kubernetes.Clientset
	config     *rest.Config
	srcSecret  string
	destSecret string
	incluster  bool
	kubeconfig string
	err        error
)

func init() {
	flag.StringVar(&srcSecret, "src-secret", "", "source secret")
	flag.StringVar(&destSecret, "dest-secret", "", "destination secret")
	flag.BoolVar(&incluster, "incluster", false, "run inside or ousite kubernetes cluster")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig location")
	flag.Parse()
}

func main() {

	if srcSecret == "" {
		fmt.Println("--src-secret is missing")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if destSecret == "" {
		fmt.Println("--dest-secret is missing")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if !incluster && kubeconfig == "" {
		fmt.Println("--kubeconfig is missing")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// setup kubenetes client config
	if incluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// stop signal for the informer
	stopper := make(chan struct{})
	defer close(stopper)

	// setup shared informers
	factory := informers.NewSharedInformerFactory(client, 0)
	secretInformer := factory.Core().V1().Secrets().Informer()

	// handle runtime crash
	defer runtime.HandleCrash()

	// start informer ->
	go factory.Start(stopper)

	// start to sync and call list
	if !cache.WaitForCacheSync(stopper, secretInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for secretInformer caches to sync"))
		return
	}

	// add handler for secret event
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    secretOnAdd,
		UpdateFunc: secretOnUpdate,
		DeleteFunc: secretOnDelete,
	})

	<-stopper
}
