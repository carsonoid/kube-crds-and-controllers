// This is functionally identical to the single-config configmap based controller
// However it reads it's configmap style reads all the keys of the configmap into a map of configs

package main // import "github.com/carsonoid/kube-crds-and-controllers/hard-coded-controller"

import (
	"encoding/json"
	"flag"
	logging "log"
	"os"
	"path/filepath"

	// Better yaml handling
	"github.com/ghodss/yaml"

	// Kubernetes and client-go
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	// "k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	log = logging.New(os.Stdout, "", logging.Lshortfile)
)

// These could be from flags if desired
const (
	configNamespace string = "kube-system"
	configName      string = "pod-labeler-config"
)

// PodLabelConfig holds the namespace to target and labels to be ensured
type PodLabelConfig struct {
	TargetNamespace string            `json:"targetNamespace"`
	Labels          map[string]string `json:"labels"`
}

// PodLabelController with a config and client
type PodLabelController struct {
	client         *kubernetes.Clientset
	Configs        map[string]*PodLabelConfig
	configLoadChan chan bool
}

// NewPodLabelController takes a kubernetes clientset and configuration and returns a valid PodLabelController
func NewPodLabelController(client *kubernetes.Clientset) *PodLabelController {
	return &PodLabelController{
		client:         client,
		configLoadChan: make(chan bool),
	}
}

// Run starts the PodLabelController and blocks until killed
func (plc *PodLabelController) Run() {
	// Start configmap watcher
	go plc.WatchConfigMap()

	log.Print("Waiting for initial config load")

	// wait for load/reload signal before starting pod controller
	<-plc.configLoadChan

	// Watch for config reloads and then restart the controller so all existing pods
	// are re-evaluted with new config
	for {
		restClient := plc.client.CoreV1().RESTClient()
		listwatch := cache.NewListWatchFromClient(restClient, "pods", corev1.NamespaceAll, fields.Everything())

		_, controller := cache.NewInformer(listwatch, &corev1.Pod{}, 0,
			cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					log.Print("Pod Add Event")
					if err := plc.handlePod(obj.(*corev1.Pod)); err != nil {
						log.Printf("Error handling pod: %s", err)
					}
				},
				UpdateFunc: func(oldobj interface{}, newobj interface{}) {
					log.Print("Pod Update Event")
					if err := plc.handlePod(newobj.(*corev1.Pod)); err != nil {
						log.Printf("Error handling pod: %s", err)
					}
				},
				DeleteFunc: func(obj interface{}) {
					log.Print("Pod Delete Event")
					// nothing to do
				},
			},
		)

		log.Print("Starting Controller")
		stopChan := make(chan struct{})
		go controller.Run(stopChan)
		<-plc.configLoadChan
		log.Print("killing controller")
		close(stopChan)
	}
}

func (plc *PodLabelController) handlePod(pod *corev1.Pod) error {
	o, err := runtime.NewScheme().DeepCopy(pod)
	if err != nil {
		return err
	}
	newPod := o.(*corev1.Pod)

	// apply labels if needed
	// if no changes then return
	if !plc.labelPod(newPod) {
		return nil
	}

	oldData, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	newData, err := json.Marshal(newPod)
	if err != nil {
		return err
	}

	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Pod{})
	if err != nil {
		return err
	}

	_, err = plc.client.CoreV1().Pods(pod.Namespace).Patch(pod.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	return nil
}

func (plc *PodLabelController) labelPod(pod *corev1.Pod) bool {
	changed := false
	// make sure map is initialized
	if len(pod.GetLabels()) == 0 {
		pod.ObjectMeta.Labels = make(map[string]string)
	}

	// Loop all configs
	for _, c := range plc.Configs {
		// only apply labels if namespace matches
		if pod.GetNamespace() == c.TargetNamespace {
			// check keys
			for k, newVal := range c.Labels {
				if curVal, ok := pod.GetLabels()[k]; ok && curVal == newVal {
					//log.Printf("Pod %s already has label: %s=%s", pod.GetName(), k, newVal)
				} else {
					log.Printf("Pod %s needs label: %s=%s", pod.GetName(), k, newVal)
					pod.Labels[k] = newVal
					changed = true
				}
			}
		}
	}
	return changed
}

func (plc *PodLabelController) WatchConfigMap() {
	log.Print("Watching for ConfigMap")

	restClient := plc.client.CoreV1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "configmaps", configNamespace, fields.OneTermEqualSelector("metadata.name", configName))

	_, controller := cache.NewInformer(listwatch, &corev1.ConfigMap{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Print("ConfigMap Add Event")
				if err := plc.loadConfigMap(obj.(*corev1.ConfigMap)); err != nil {
					log.Printf("Error loading config from configmap: %s", err)
				}
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				log.Print("ConfigMap Update Event")
				if err := plc.loadConfigMap(newobj.(*corev1.ConfigMap)); err != nil {
					log.Printf("Error loading config from configmap: %s", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Print("ConfigMap Deleted - last known config will be retained")
				// nothing to do
			},
		},
	)

	stopChan := make(chan struct{})
	controller.Run(stopChan)
	<-stopChan
}

func (plc *PodLabelController) loadConfigMap(cm *corev1.ConfigMap) error {
	log.Print("Loading ConfigMap")

	// make a new map so deleted keys get removed
	plc.Configs = make(map[string]*PodLabelConfig)

	// Load all config keys from the map as configurations
	for k, v := range cm.Data {
		// New empty config struct
		c := PodLabelConfig{}

		// Populate struct from value
		if err := yaml.Unmarshal([]byte(v), &c); err != nil {
			return err
		}

		// Update map with pointer to struct
		plc.Configs[k] = &c

	}

	log.Printf("Loaded %d new configs", len(plc.Configs))

	// Send Load/Reload signal
	plc.configLoadChan <- true

	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create controller, passing only the kube client
	plc := NewPodLabelController(clientset)

	// Run controller
	plc.Run()
}
