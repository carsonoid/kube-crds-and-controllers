// This is essentially the "crd-configured/workqueue" controller from this repo.
// But with a different resource and different business logic

package main

import (
	"bytes"
	"flag"
	"fmt"
	logging "log"
	"os"
	"path/filepath"
	"text/template"
	"time"

	// Kubernetes and client-go
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	machinery_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"

	// Custom resources
	wpv1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/apis/provisioner/v1alpha1"
	wpclient "github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/client/clientset/versioned"
)

// BONUS: These values could be read dynamically from a configmap
const NamespacePrefix string = "wa-"
const AttendeeServiceAccountName string = "attendee"
const AttendeeFinalizer string = "workshopattendee.finalizers.k8s.carsonoid.net"
const AttendeeServiceAccountClusterRoleName string = "podlabeler"
const KubeconfigTemplate string = `apiVersion: v1
kind: Config
preferences: {}
current-context: "workshop"
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: {{ .clusteraddr }}
  name: workshop
contexts:
- context:
    cluster: workshop
    namespace: {{ .namespace }}
    user: workshop
  name: workshop
users:
- name: workshop
  user:
    token: {{ .token }}
`

var (
	log = logging.New(os.Stdout, "", logging.Lshortfile)
)

// WorkshopProvisionerController with a config and client
type WorkshopProvisionerController struct {
	client      *kubernetes.Clientset
	wpClientset *wpclient.Clientset

	numAttendeeWorkers *int
	ClusterAddr        *string

	attendeeIndexer  cache.Indexer
	attendeeQueue    workqueue.RateLimitingInterface
	attendeeInformer cache.Controller
}

// NewWorkshopProvisionerController takes a kubernetes clientset and configuration and returns a valid WorkshopProvisionerController
func NewWorkshopProvisionerController(client *kubernetes.Clientset, wpClientset *wpclient.Clientset, numAttendeeWorkers *int, ca *string) *WorkshopProvisionerController {
	return &WorkshopProvisionerController{
		client:             client,
		wpClientset:        wpClientset,
		numAttendeeWorkers: numAttendeeWorkers,
		ClusterAddr:        ca,
	}
}

// Run starts the WorkshopProvisionerController and blocks until killed
func (wpc *WorkshopProvisionerController) Run() {
	killChan := make(chan struct{})
	defer close(killChan)

	// Start crd controller
	go wpc.StartAttendeeController(killChan)

	// BONUS: The various sub-resources could be watched for changes/deletes directly. And reconciles could be immediately triggered.

	<-killChan
}

func (wpc *WorkshopProvisionerController) StartAttendeeController(killChan chan struct{}) {
	log.Println("Starting WorkshopAttendee controller")

	restClient := wpc.wpClientset.ProvisionerV1alpha1().RESTClient()
	listwatch := cache.NewListWatchFromClient(restClient, "workshopattendees", corev1.NamespaceAll, fields.Everything())

	wpc.attendeeQueue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	wpc.attendeeIndexer, wpc.attendeeInformer = cache.NewIndexerInformer(listwatch, &wpv1alpha1.WorkshopAttendee{}, time.Second*30,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				// log.Print("WorkshopAttendee Add Event")
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					wpc.attendeeQueue.Add(key)
				}
			},
			UpdateFunc: func(oldobj interface{}, newobj interface{}) {
				// log.Print("WorkshopAttendee Update Event")
				key, err := cache.MetaNamespaceKeyFunc(newobj)
				if err == nil {
					wpc.attendeeQueue.Add(key)
				}
			},
			DeleteFunc: func(obj interface{}) {
				// log.Print("WorkshopAttendee Delete Event")
				// Rester delete with delta queue
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					wpc.attendeeQueue.Add(key)
				}
			},
		}, cache.Indexers{})

	go wpc.StartQueueWorkers(*wpc.numAttendeeWorkers, killChan)
}

func (wpc *WorkshopProvisionerController) StartQueueWorkers(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer wpc.attendeeQueue.ShutDown()
	log.Println("Starting Attendee Queue Workers")

	go wpc.attendeeInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, wpc.attendeeInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(wpc.runQueueWorker, time.Second, stopCh)
	}

	<-stopCh
	log.Println("Stopping Attendee Queue Workers")
}

func (wpc *WorkshopProvisionerController) runQueueWorker() {
	for wpc.processNextPod() {
	}
}

func (wpc *WorkshopProvisionerController) processNextPod() bool {
	// Wait until there is a new item in the working queue
	key, quit := wpc.attendeeQueue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two resources with the same key are never processed in
	// parallel.
	defer wpc.attendeeQueue.Done(key)

	// Invoke the method containing the business logic
	err := wpc.processAttendee(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	wpc.handleErr(err, key)
	return true
}

func (wpc *WorkshopProvisionerController) processAttendee(key string) error {
	obj, exists, err := wpc.attendeeIndexer.GetByKey(key)
	if err != nil {
		log.Printf("Fetching object with key %s from store failed with %v\n", key, err)
		return err
	}

	if exists {
		wpc.reconcileAttendee(obj.(*wpv1alpha1.WorkshopAttendee))
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (wpc *WorkshopProvisionerController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		wpc.attendeeQueue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if wpc.attendeeQueue.NumRequeues(key) < 5 {
		log.Printf("Error syncing pod %v: %v\n", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		wpc.attendeeQueue.AddRateLimited(key)
		return
	}

	wpc.attendeeQueue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	log.Printf("Dropping pod %q out of the queue: %v\n", key, err)
}

func (wpc *WorkshopProvisionerController) reportChange(resource string, name string) error {
	log.Printf("Reporting Change for %s, %s", resource, name)
	return nil
}

func (wpc *WorkshopProvisionerController) GetNamespaceName(wa *wpv1alpha1.WorkshopAttendee) string {
	return NamespacePrefix + wa.GetName()
}

func (wpc *WorkshopProvisionerController) GetServiceAccountToken(wa *wpv1alpha1.WorkshopAttendee) (string, error) {
	saClient := wpc.client.CoreV1().ServiceAccounts(wpc.GetNamespaceName(wa))

	// Get the
	sa, saErr := saClient.Get(AttendeeServiceAccountName, metav1.GetOptions{})
	if saErr != nil {
		log.Printf("Error getting service account token for %s", wa.GetName())
		return "", saErr
	}

	secretClient := wpc.client.CoreV1().Secrets(wpc.GetNamespaceName(wa))

	// Create the namespace if it doesn't exist
	secret, secretErr := secretClient.Get(sa.Secrets[0].Name, metav1.GetOptions{})
	if secretErr != nil {
		log.Printf("Error getting secret token for %s", wa.GetName())
		return "", secretErr
	}

	return string(secret.Data["token"]), nil
}

func (wpc *WorkshopProvisionerController) GetKubeconfig(wa *wpv1alpha1.WorkshopAttendee) string {
	// set template context vars
	vars := make(map[string]string)
	vars["clusteraddr"] = *wpc.ClusterAddr
	vars["namespace"] = wpc.GetNamespaceName(wa)
	token, _ := wpc.GetServiceAccountToken(wa)
	vars["token"] = token

	tmpl, err := template.New("kubeconfig").Parse(KubeconfigTemplate)
	if err != nil {
		log.Printf("%s", err)
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, vars)
	if err != nil {
		log.Printf("%s", err)
	}
	return out.String()
}

func (wpc *WorkshopProvisionerController) UpdateChildStatus(wa *wpv1alpha1.WorkshopAttendee, resource string) error {
	log.Printf("Setting child status for %s:%s", wa.GetName(), resource)

	provisionerClient := wpc.wpClientset.ProvisionerV1alpha1().WorkshopAttendees()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the WorkshopAttendee before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := provisionerClient.Get(wa.GetName(), metav1.GetOptions{})
		if getErr != nil {
			log.Printf("Failed to get latest version of WorkshopAttendee: %v", getErr)
		}

		// initialize empty map if needed
		if result.Status.Children == nil {
			result.Status.Children = make(map[string]metav1.Time)
		}

		result.Status.Children[resource] = metav1.Now()

		// If we are setting a child status, We are assumed to be in a creating state again.
		result.Status.State = wpv1alpha1.WorkshopAttendeeStateCreating

		_, updateErr := provisionerClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func (wpc *WorkshopProvisionerController) UpdateFinalState(wa *wpv1alpha1.WorkshopAttendee) error {
	log.Printf("Updating final state for %s", wa.GetName())

	provisionerClient := wpc.wpClientset.ProvisionerV1alpha1().WorkshopAttendees()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the WorkshopAttendee before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := provisionerClient.Get(wa.GetName(), metav1.GetOptions{})
		if getErr != nil {
			log.Printf("Failed to get latest version of WorkshopAttendee: %v", getErr)
		}

		// Set ready state
		result.Status.State = wpv1alpha1.WorkshopAttendeeStateReady

		// set kubeconfig
		result.Status.Kubeconfig = wpc.GetKubeconfig(wa)

		_, updateErr := provisionerClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func (wpc *WorkshopProvisionerController) UpdateState(wa *wpv1alpha1.WorkshopAttendee, s wpv1alpha1.WorkshopAttendeeState) error {
	log.Printf("Updating state for %s:%s", wa.GetName(), s)

	provisionerClient := wpc.wpClientset.ProvisionerV1alpha1().WorkshopAttendees()

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of the WorkshopAttendee before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := provisionerClient.Get(wa.GetName(), metav1.GetOptions{})
		if getErr != nil {
			log.Printf("Failed to get latest version of WorkshopAttendee: %v", getErr)
		}

		result.Status.State = s

		_, updateErr := provisionerClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		return retryErr
	}
	return nil
}

func filter(vs []string, f func(string) bool) []string {
	var vsf []string
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func (wpc *WorkshopProvisionerController) removeFinalizer(wa *wpv1alpha1.WorkshopAttendee) error {
	provisionerClient := wpc.wpClientset.ProvisionerV1alpha1().WorkshopAttendees()

	// Finalizer not already set. Add it
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := provisionerClient.Get(wa.GetName(), metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		// Create filtered finalizers list, remove completed finalizer
		newFinalizers := filter(result.GetFinalizers(), func(v string) bool {
			if AttendeeFinalizer == v {
				return false
			}
			return true
		})

		// Set new list of finalizers on obj and update
		result.SetFinalizers(newFinalizers)

		_, updateErr := provisionerClient.Update(result)
		return updateErr
	})

	if retryErr != nil {
		log.Printf("Update failed: %+v", retryErr)
		return retryErr
	}

	log.Printf("Removed finalizer from: %s", wa.GetName())

	return nil
}

func (wpc *WorkshopProvisionerController) reconcileFinalizer(wa *wpv1alpha1.WorkshopAttendee) error {
	provisionerClient := wpc.wpClientset.ProvisionerV1alpha1().WorkshopAttendees()

	// Only add finalizer if it's not already present
	for _, f := range wa.GetFinalizers() {
		if f == AttendeeFinalizer {
			return nil
		}
	}

	// Finalizer not already set. Add it
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := provisionerClient.Get(wa.GetName(), metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		// Add Finalizer
		result.SetFinalizers(append(result.GetFinalizers(), AttendeeFinalizer))

		_, updateErr := provisionerClient.Update(result)
		return updateErr
	})

	if retryErr != nil {
		log.Printf("Update failed: %+v", retryErr)
		return retryErr
	}

	log.Printf("Added finalizer for: %s", wa.GetName())
	return nil
}

func (wpc *WorkshopProvisionerController) reconcileNamespace(wa *wpv1alpha1.WorkshopAttendee) error {
	nsName := wpc.GetNamespaceName(wa)
	nsClient := wpc.client.CoreV1().Namespaces()

	// Create the namespace if it doesn't exist
	if _, err := nsClient.Get(nsName, metav1.GetOptions{}); err != nil {
		log.Printf("Creating Namespace for %s", wa.GetName())
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: nsName,
			},
		}

		_, err := nsClient.Create(ns)
		if err != nil {
			log.Printf("Error creating namespace: %s", err)
			return err
		}

		wpc.UpdateChildStatus(wa, "namespace")
	}

	return nil
}

func (wpc *WorkshopProvisionerController) reconcileServiceAccount(wa *wpv1alpha1.WorkshopAttendee) error {
	saClient := wpc.client.CoreV1().ServiceAccounts(wpc.GetNamespaceName(wa))

	// Create the namespace if it doesn't exist
	if _, err := saClient.Get(AttendeeServiceAccountName, metav1.GetOptions{}); err != nil {
		log.Printf("Creating ServiceAccount for %s", wa.GetName())
		ns := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: AttendeeServiceAccountName,
			},
		}

		_, err := saClient.Create(ns)
		if err != nil {
			log.Printf("Error creating namespace: %s", err)
			return err
		}

		wpc.UpdateChildStatus(wa, "serviceaccount")
	}

	return nil
}

func (wpc *WorkshopProvisionerController) reconcileRoleBinding(wa *wpv1alpha1.WorkshopAttendee) error {
	nsName := wpc.GetNamespaceName(wa)
	rbClient := wpc.client.RbacV1beta1().RoleBindings(nsName)

	// Create the namespace if it doesn't exist
	if _, err := rbClient.Get(AttendeeServiceAccountName, metav1.GetOptions{}); err != nil {
		log.Printf("Creating RoleBinding for %s", wa.GetName())
		ns := &rbacv1beta1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      AttendeeServiceAccountName,
				Namespace: nsName,
			},
			RoleRef: rbacv1beta1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "podlabeler",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      AttendeeServiceAccountName,
					Namespace: nsName,
				},
			},
		}

		_, err := rbClient.Create(ns)
		if err != nil {
			log.Printf("Error creating namespace: %s", err)
			return err
		}

		wpc.UpdateChildStatus(wa, "rolebinding")
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }

func (wpc *WorkshopProvisionerController) reconcileDeployments(wa *wpv1alpha1.WorkshopAttendee) error {
	nsName := wpc.GetNamespaceName(wa)
	depClient := wpc.client.AppsV1beta2().Deployments(nsName)

	apps := []string{"app1", "app2", "app3"}

	// Create the deployments if they don't exist
	for _, app := range apps {
		if _, err := depClient.Get(app, metav1.GetOptions{}); err != nil {
			log.Printf("Creating Deployment %s for %s", app, wa.GetName())
			deployment := &appsv1beta2.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: app,
				},
				Spec: appsv1beta2.DeploymentSpec{
					Replicas: int32Ptr(2),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": app,
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": app,
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  app,
									Image: "nginx:1.12",
									Ports: []corev1.ContainerPort{
										{
											Name:          "http",
											Protocol:      corev1.ProtocolTCP,
											ContainerPort: 80,
										},
									},
								},
							},
						},
					},
				},
			}

			_, err := depClient.Create(deployment)
			if err != nil {
				log.Printf("Error creating deployment: %s", err)
				return err
			}

			wpc.UpdateChildStatus(wa, "deployment:"+app)
		}
	}

	return nil
}

func (wpc *WorkshopProvisionerController) reportOnReady(wa *wpv1alpha1.WorkshopAttendee) error {
	// Report changes when moving from Creating -> Ready Status
	if wa.Status.State == wpv1alpha1.WorkshopAttendeeStateCreating {
		log.Printf("WorkshopAttendee %s provisioning is completed or a resource has been recreated", wa.GetName())

		// Update status/kubeconfig
		wpc.UpdateFinalState(wa)

		// Send result email
		log.Printf("Sent email to %s", wa.Spec.Email)
	}

	return nil
}

func (wpc *WorkshopProvisionerController) deleteAttendeeResources(wa *wpv1alpha1.WorkshopAttendee) error {
	// Check/Delete Namespace. Let the namespace controller do the hard work
	nsName := wpc.GetNamespaceName(wa)
	nsClient := wpc.client.CoreV1().Namespaces()

	// Deleting the namespace will trigger it's teardown process. We keep checking/deleting until the resource is gone
	log.Printf("Deleting all resources for %s", wa.GetName())
	for {
		// Delete the namespace it exists and is not already set to be deleted
		r, err := nsClient.Get(nsName, metav1.GetOptions{})

		// namespace already deleted
		if errors.IsNotFound(err) {
			log.Printf("Namespace for %s is deleted", wa.GetName())
			break
		}

		// NS exists and is not already set for deletion
		if err == nil && r.GetDeletionTimestamp() == nil {
			err := nsClient.Delete(nsName, &metav1.DeleteOptions{})

			if err != nil {
				log.Printf("Error deleting namespace: %s", err)
				return err
			}

			wpc.UpdateState(wa, wpv1alpha1.WorkshopAttendeeStateDeleting)
		}

		log.Printf("Waiting for %s delete", wa.GetName())
		time.Sleep(time.Second * 3)
	}

	log.Printf("Resources for attendee %s deleted", wa.GetName())
	wpc.removeFinalizer(wa)
	return nil
}

func (wpc *WorkshopProvisionerController) reconcileAttendee(in *wpv1alpha1.WorkshopAttendee) error {
	o, err := machinery_runtime.NewScheme().DeepCopy(in)
	if err != nil {
		return err
	}
	wa := o.(*wpv1alpha1.WorkshopAttendee)

	// If the DeleteTimestamp is set. Do the delete instead.
	if wa.GetDeletionTimestamp() != nil {
		wpc.UpdateState(wa, wpv1alpha1.WorkshopAttendeeStateDeleting)
		return wpc.deleteAttendeeResources(wa)
	}

	log.Printf("Reconciling Attendee: %s", wa.GetName())

	// Mark new attendees as creating
	if wa.Status.State == "" {
		log.Printf("WorkshopAttendee %s provisioning is done", wa.GetName())

		// update status
		wpc.UpdateState(wa, wpv1alpha1.WorkshopAttendeeStateCreating)
	}

	// BONUS: Mostof these actions are very similar. There could be some cleanup to remove repitition

	// Make sure the finalizer is installed
	wpc.reconcileFinalizer(wa)

	// Reconcile Namespace
	wpc.reconcileNamespace(wa)

	// Reconcile ServiceAccount
	wpc.reconcileServiceAccount(wa)

	// Reconcile RoleBinding
	wpc.reconcileRoleBinding(wa)

	// Reconcile Deloyments
	wpc.reconcileDeployments(wa)

	// Finalize
	wpc.reportOnReady(wa)

	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	var clusterAddr *string
	clusterAddr = flag.String("cluster-addr", "https://kubernetes", "cluster address for generated kubeconfig")

	var numAttendeeWorkers *int
	numAttendeeWorkers = flag.Int("num-attendee-workers", 5, "(optional) number of concurrent attendee workers")
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
	wpClientset, err := wpclient.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// BONUS: A separate controller could be created/ran which constantly ensures all the cluster resources required to run the controller (clusterrole, etcs)

	// Create controller, passing all clients
	wpc := NewWorkshopProvisionerController(clientset, wpClientset, numAttendeeWorkers, clusterAddr)

	// Run controller
	wpc.Run()
}
