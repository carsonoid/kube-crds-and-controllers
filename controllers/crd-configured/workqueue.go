// This is functionally identical to the multi-config configmap based controller
// However it uses a cluster-scoped kubernetes resource to get it's configurations

package main // import "github.com/carsonoid/kube-crds-and-controllers/hard-coded-controller"

import (
	"encoding/json"
	"flag"
	"fmt"
	logging "log"
	"os"
	"path/filepath"
	"time"

	// Kubernetes and client-go
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	machinery_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	// Custom resources
	plv1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis/podlabeler/v1alpha1"
	plclient "github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/client/clientset/versioned"
)

var (
	log = logging.New(os.Stdout, "", logging.Lshortfile)
)

// PodLabelController with a config and client
type PodLabelController struct {
	client      *kubernetes.Clientset
	plClientset *plclient.Clientset
	HasSynced   bool

	podLabelConfigStore      cache.Store
	podLabelConfigController cache.Controller

	numPodWorkers *int
	podIndexer    cache.Indexer
	podQueue      workqueue.RateLimitingInterface
	podInformer   cache.Controller
}

// NewPodLabelController takes a kubernetes clientset and configuration and returns a valid PodLabelController
func NewPodLabelController(client *kubernetes.Clientset, plClientset *plclient.Clientset, numPodWorkers *int) *PodLabelController {
	return &PodLabelController{
		client:        client,
		plClientset:   plClientset,
		numPodWorkers: numPodWorkers,
	}
}

// Run starts the PodLabelController and blocks until killed
func (plc *PodLabelController) Run() {
	killChan := make(chan struct{})
	defer close(killChan)

	// Start watching PodLabelConfigs
	plc.StartPodLabelConfigController(killChan)

	log.Print("Waiting for initial PodLabelConfig sync")

	// Wait for store to sync up before processing pods
	if !cache.WaitForCacheSync(killChan, plc.podLabelConfigController.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	plc.HasSynced = true

	log.Print("Initial PodLabelConfig sync complete")

	// Start pod controller
	go plc.StartPodController(killChan)
	<-killChan
}

func (plc *PodLabelController) StartPodController(killChan chan struct{}) {
	log.Println("Starting Pod controller")

	restClient := plc.client.CoreV1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "pods", corev1.NamespaceAll, fields.Everything())

	plc.podQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	plc.podIndexer, plc.podInformer = cache.NewIndexerInformer(listwatch, &corev1.Pod{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Print("Pod Add Event")
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					plc.podQueue.Add(key)
				}
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				log.Print("Pod Update Event")
				// Make sure object is not set for deltion and was actually changed
				if newobj.(*corev1.Pod).GetDeletionTimestamp() == nil &&
					oldobj.(*corev1.Pod).GetResourceVersion() != newobj.(*corev1.Pod).GetResourceVersion() {
					key, err := cache.MetaNamespaceKeyFunc(newobj)
					if err == nil {
						plc.podQueue.Add(key)
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Print("Pod Delete Event")
				// Rester delete with delta queue
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					plc.podQueue.Add(key)
				}
			},
		}, cache.Indexers{})

	// Watch for config reloads and then restart the controller so all existing pods
	// are re-evaluted with new config

	go plc.StartQueueWorkers(*plc.numPodWorkers, killChan)
}

func (plc *PodLabelController) StartQueueWorkers(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer plc.podQueue.ShutDown()
	log.Println("Starting Pod Queue Workers")

	go plc.podInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, plc.podInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(plc.runPodQueueWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Println("Stopping Pod Queue Workers")
}

func (plc *PodLabelController) runPodQueueWorker() {
	for plc.processNextPod() {
	}
}

func (plc *PodLabelController) processNextPod() bool {
	// Wait until there is a new item in the working queue
	key, quit := plc.podQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer plc.podQueue.Done(key)

	// Invoke the method containing the business logic
	err := plc.processPod(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	plc.handleErr(err, key)
	return true
}

func (plc *PodLabelController) processPod(key string) error {
	obj, exists, err := plc.podIndexer.GetByKey(key)
	if err != nil {
		log.Printf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if exists {
		plc.handlePod(obj.(*corev1.Pod))
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (plc *PodLabelController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		plc.podQueue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if plc.podQueue.NumRequeues(key) < 5 {
		log.Printf("Error syncing pod %v: %v\n", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		plc.podQueue.AddRateLimited(key)
		return
	}

	plc.podQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Printf("Dropping pod %q out of the queue: %v\n", key, err)
}

func (plc *PodLabelController) handlePod(pod *corev1.Pod) error {
	o, err := machinery_runtime.NewScheme().DeepCopy(pod)
	if err != nil {
		return err
	}
	newPod := o.(*corev1.Pod)

	// apply labels if needed
	// if no changes then return
	if !plc.labelPod(newPod) {
		return nil
	}

	// Uncomment to test threaded queue
	// log.Printf("Long operation on %s starting\n", pod.GetName())
	// time.Sleep(time.Second * 3)
	// log.Printf("Long operation on %s done\n", pod.GetName())

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
					// log.Printf("Pod %s already has label: %s=%s", pod.GetName(), k, newVal)
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

func (plc *PodLabelController) ReconcileAllPods(c *plv1alpha1.PodLabelConfig) {
	// Only reconcile after initial sync
	if !plc.HasSynced {
		return
	}

	log.Printf("Reconciling of all pods for plc: %s\n", c.GetNamespace())
	pods, err := plc.client.CoreV1().Pods(c.GetNamespace()).List(metav1.ListOptions{})
	if err != nil {
		log.Println(err)
	}
	for _, p := range pods.Items {
		if err := plc.handlePod(&p); err != nil {
			log.Printf("Error handling pod: %s", err)
		}
	}
}

func (plc *PodLabelController) StartPodLabelConfigController(killChan chan struct{}) {
	log.Print("Starting PodLabelConfig Controller")

	restClient := plc.plClientset.PodlabelerV1alpha1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "podlabelconfigs", corev1.NamespaceAll, fields.Everything())

	plc.podLabelConfigStore, plc.podLabelConfigController = cache.NewInformer(listwatch, &plv1alpha1.PodLabelConfig{}, 0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Print("PodLabelConfig Add Event")
				plc.ReconcileAllPods(obj.(*plv1alpha1.PodLabelConfig))
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				log.Print("PodLabelConfig Update Event")
				// Make sure object is not set for deltion and was actually changed
				if newobj.(*plv1alpha1.PodLabelConfig).GetDeletionTimestamp() == nil &&
					oldobj.(*plv1alpha1.PodLabelConfig).GetResourceVersion() != newobj.(*plv1alpha1.PodLabelConfig).GetResourceVersion() {
					plc.ReconcileAllPods(newobj.(*plv1alpha1.PodLabelConfig))
				}
			},
			DeleteFunc: func(obj interface{}) {
				log.Print("PodLabelConfig Delete Event")
				// Do nothing
			},
		},
	)

	go plc.podLabelConfigController.Run(killChan)
}

func main() {
	log.SetOutput(os.Stdout)

	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	var numPodWorkers *int
	numPodWorkers = flag.Int("num-pod-workers", 1, "(optional) number of concurrent pod workers")
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
	plc := NewPodLabelController(clientset, plClientset, numPodWorkers)

	// Run controller
	plc.Run()
}
