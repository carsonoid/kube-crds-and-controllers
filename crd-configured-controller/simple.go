// This is functionally identical to the multi-config configmap based controller
// However it uses a cluster-scoped kubernetes resource to get it's configurations

package main // import "github.com/carsonoid/kube-crds-and-controllers/hard-coded-controller"

import (
	"encoding/json"
	"flag"
	logging "log"
	"os"
	"path/filepath"

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

	// Custom resources
	plv1alpha1 "github.com/carsonoid/kube-crds-and-controllers/crd-configured-controller/pkg/apis/podlabeler/v1alpha1"
	plclient "github.com/carsonoid/kube-crds-and-controllers/crd-configured-controller/pkg/client/clientset/versioned"
)

var (
	log = logging.New(os.Stdout, "", logging.Lshortfile)
)

// PodLabelController with a config and client
type PodLabelController struct {
	client                   *kubernetes.Clientset
	plClientset              *plclient.Clientset
	podLabelConfigStore      cache.Store
	podLabelConfigController cache.Controller
	configLoadChan           chan bool
}

// NewPodLabelController takes a kubernetes clientset and configuration and returns a valid PodLabelController
func NewPodLabelController(client *kubernetes.Clientset, plClientset *plclient.Clientset) *PodLabelController {
	return &PodLabelController{
		client:         client,
		plClientset:    plClientset,
		configLoadChan: make(chan bool),
	}
}

// Run starts the PodLabelController and blocks until killed
func (plc *PodLabelController) Run() {
	killChan := make(chan struct{})
	defer close(killChan)
	go plc.WatchCustomResources(killChan)
	go plc.StartPodController(killChan)
	<-killChan
}

func (plc *PodLabelController) StartPodController(killChan chan struct{}) {
	log.Print("Waiting for initial customresource sync")

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

	// Watch for config reloads and then restart the controller so all existing pods
	// are re-evaluted with new config
	for {
		log.Print("Starting Controller")
		stopChan := make(chan struct{})
		go controller.Run(stopChan)
		<-plc.configLoadChan
		log.Print("killing controller")
		close(stopChan)
	}
	<-killChan
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
	for _, obj := range plc.podLabelConfigStore.List() {
		c := obj.(*plv1alpha1.PodLabelConfig)
		// only apply labels if namespace matches
		if pod.GetNamespace() == c.GetNamespace() {
			// check keys
			for k, newVal := range c.Spec.Labels {
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

func (plc *PodLabelController) WatchCustomResources(stopChan chan struct{}) {
	log.Print("Watching for CustomResources")

	restClient := plc.plClientset.PodlabelerV1alpha1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "podlabelconfigs", corev1.NamespaceAll, fields.Everything())

	plc.podLabelConfigStore, plc.podLabelConfigController = cache.NewInformer(listwatch, &plv1alpha1.PodLabelConfig{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Print("PodLabelConfig Add Event")
				plc.configLoadChan <- true
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				log.Print("PodLabelConfig Update Event")
				// make sure object was actually changed
				if oldobj.(*plv1alpha1.PodLabelConfig).GetResourceVersion() != newobj.(*plv1alpha1.PodLabelConfig).GetResourceVersion() {
					plc.configLoadChan <- true
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Print("PodLabelConfig Deleted Event")
				// nothing to do
			},
		},
	)

	go plc.podLabelConfigController.Run(stopChan)
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

	// create the clientset
	plClientset, err := plclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create controller, passing all clients
	plc := NewPodLabelController(clientset, plClientset)

	// Run controller
	plc.Run()
}
