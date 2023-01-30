package cmd

import (
	"fmt"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	err     error

	client *kubernetes.Clientset
	config *rest.Config

	srcSecret  string
	destSecret string
	incluster  bool
	kubeconfig string
)

var rootCmd = &cobra.Command{
	Use:   "config-syncer",
	Short: "synchronize kubernetes configmaps/secrets between namespaces",
	Long:  "Config Syncer is a utility to synchronize specified Kubernetes secrets between different namespaces",
	Run: func(cmd *cobra.Command, args []string) {

		// get kubeconfig, either fron kubeconfig arg or incluster kubeconfig
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

		// cteate new kubernetes client
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

		// handle stop signal
		<-stopper
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// setup things when cobra initiate
	cobra.OnInitialize(initConfig)

	// global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config-syncer.yaml")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug log message")

	// local flags
	rootCmd.Flags().StringVar(&srcSecret, "src-secret", "", "source secret")
	rootCmd.Flags().StringVar(&destSecret, "dest-secret", "", "destination secret")
	rootCmd.Flags().BoolVar(&incluster, "incluster", false, "run inside or ousite kubernetes cluster")
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig location")

	// validate required flags
	rootCmd.MarkFlagRequired("src-secret")
	rootCmd.MarkFlagRequired("dest-secret")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// validate kubeconfig must exists when running outside cluster
	if !incluster {
		rootCmd.MarkFlagRequired("kubeconfig")
	}

	if cfgFile != "" {
		// use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// search config in current directory with name "config-syncer.yaml".
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config-syncer")
	}

	// override configuration through viper
	viper.Set("debug", debug)
	if debug {
		log.Println("running in DEBUG mode")
	}

	// if a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "using config file:", viper.ConfigFileUsed())
	}
}
