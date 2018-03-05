// This is functionally identical to the simple hard-coded controller.
// But doesn't use global variables for configuration

package main // import "github.com/carsonoid/kube-crds-and-controllers/hard-coded-controller"

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	logging "log"
	"os"
	"path/filepath"

	// Better yaml handling
	"github.com/ghodss/yaml"

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

// PodLabelConfig holds the namespace to target and labels to be ensured
type PodLabelConfig struct {
	TargetNamespace string            `json:"targetNamespace"`
	Labels          map[string]string `json:"labels"`
}

// PodLabelController with a config and client
type PodLabelController struct {
	client *kubernetes.Clientset
	Config *PodLabelConfig
}

// NewPodLabelController takes a kubernetes clientset and configuration and returns a valid PodLabelController
func NewPodLabelController(client *kubernetes.Clientset, config *PodLabelConfig) *PodLabelController {
	return &PodLabelController{
		client: client,
		Config: config,
	}
}

// Run starts the PodLabelController and blocks until killed
func (plc *PodLabelController) Run() {
	log.Print("Starting Controller")

	restClient := plc.client.CoreV1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "pods", plc.Config.TargetNamespace, fields.Everything())

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

	stopChan := make(chan struct{})
	go controller.Run(stopChan)
	<-stopChan
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

	// check keys
	for k, newVal := range plc.Config.Labels {
		if curVal, ok := pod.GetLabels()[k]; ok && curVal == newVal {
			//log.Printf("Pod %s already has label: %s=%s", pod.GetName(), k, newVal)
		} else {
			log.Printf("Pod %s needs label: %s=%s", pod.GetName(), k, newVal)
			pod.Labels[k] = newVal
			changed = true
		}
	}
	return changed
}

func main() {
	log.SetOutput(os.Stdout)

	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	var configPath *string
	configPath = flag.String("config", "", "(optional) custom PodLabelConfig file")
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

	// Load default config or one from a file
	var podlabelconfig *PodLabelConfig
	if *configPath == "" {
		podlabelconfig = &PodLabelConfig{
			TargetNamespace: "default",
			Labels: map[string]string{
				"is-from-structured": "true",
			},
		}
	} else {
		y, err := ioutil.ReadFile(*configPath)
		if err != nil {
			log.Printf("Error reading config file: %v\n", err)
			return
		}
		err = yaml.Unmarshal(y, &podlabelconfig)
		if err != nil {
			log.Printf("Error unmarshaling config yaml: %v\n", err)
			return
		}
	}

	// Run controller with hard-coded config
	plc := NewPodLabelController(clientset, podlabelconfig)

	plc.Run()
}
