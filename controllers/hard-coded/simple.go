// This controller watches for all pods in a configured namespace and makes sure that they
// all have a hard-coded set of labels. It will add them to existing pods, new pods,
// and will ensure that they have the labels on every update

package main // import "github.com/carsonoid/kube-crds-and-controllers/hard-coded-controller"

import (
	"encoding/json"
	"flag"
	logging "log"
	"os"
	"path/filepath"

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

// "hard-coded" default holders
var labels map[string]string
var targetNamespace string

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

	// Namespace
	targetNamespace = "default"

	// initialize labels
	labels = make(map[string]string)
	labels["is-awesome"] = "true"
	labels["is-not-awesome"] = "false"

	runController(clientset)
}

func runController(client *kubernetes.Clientset) {
	log.Print("Starting Controller")

	restClient := client.CoreV1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "pods", targetNamespace, fields.Everything())

	_, controller := cache.NewInformer(listwatch, &corev1.Pod{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Print("Pod Add Event")
				if err := handlePod(obj.(*corev1.Pod), client); err != nil {
					log.Printf("Error handling pod: %s", err)
				}
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				log.Print("Pod Update Event")
				if err := handlePod(newobj.(*corev1.Pod), client); err != nil {
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

func handlePod(pod *corev1.Pod, client *kubernetes.Clientset) error {
	o, err := runtime.NewScheme().DeepCopy(pod)
	if err != nil {
		return err
	}
	newPod := o.(*corev1.Pod)

	// apply labels if needed
	// if no changes then return
	if !labelPod(newPod) {
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

	_, err = client.CoreV1().Pods(pod.Namespace).Patch(pod.Name, types.StrategicMergePatchType, patchBytes)
	if err != nil {
		return err
	}

	return nil
}

func labelPod(pod *corev1.Pod) bool {
	changed := false
	// make sure map is initialized
	if len(pod.GetLabels()) == 0 {
		pod.ObjectMeta.Labels = make(map[string]string)
	}

	// check keys
	for k, newVal := range labels {
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
