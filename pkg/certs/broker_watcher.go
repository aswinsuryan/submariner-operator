package certs

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/submariner-io/admiral/pkg/certificate"
	"github.com/submariner-io/admiral/pkg/util"
	"github.com/submariner-io/submariner-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertificateControllerState manages the state of certificate controllers
type CertificateControllerState struct {
	mutex                 sync.RWMutex
	enabled               bool
	activeBrokerNamespace string
}

var certControllerState = &CertificateControllerState{}

// BrokerWatcherReconciler watches for namespace events and checks for brokers
type BrokerWatcherReconciler struct {
	client.Client
	Log               logr.Logger
	Config            *rest.Config
	OperatorNamespace string

	informerMutex           sync.Mutex
	informerStopCh          chan struct{}
	informerActiveNamespace string
	signer                  certificate.Signer
}

// No Reconcile method - a dedicated informer handles Broker CR events

func (r *BrokerWatcherReconciler) handleBrokerDeleted(brokerNamespace string) (ctrl.Result, error) {
	certControllerState.mutex.Lock()
	defer certControllerState.mutex.Unlock()

	if certControllerState.enabled && certControllerState.activeBrokerNamespace == brokerNamespace {
		r.Log.Info("Active broker deleted, disabling certificate controllers", "namespace", brokerNamespace)

		// Stop signer if running
		if r.signer != nil {
			r.signer.Stop(brokerNamespace)
		}

		certControllerState.enabled = false
		certControllerState.activeBrokerNamespace = ""
	}

	return ctrl.Result{}, nil
}

func (r *BrokerWatcherReconciler) enableControllers(brokerNamespace string, directClient client.Client, log logr.Logger) (ctrl.Result, error) {
	certControllerState.mutex.Lock()
	defer certControllerState.mutex.Unlock()

	certControllerState.enabled = true
	certControllerState.activeBrokerNamespace = brokerNamespace

	// Issue CA certificate and start CSR signer using Admiral
	log.Info("Creating CA certificate and starting CSR signer for broker namespace", "brokerNamespace", brokerNamespace)
	if err := r.setupCertificateManagement(brokerNamespace); err != nil {
		log.Error(err, "Failed to setup certificate management", "brokerNamespace", brokerNamespace)
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	log.Info("Certificate controllers enabled, CA created, and CSR signer started successfully", "brokerNamespace", brokerNamespace)
	return ctrl.Result{}, nil
}

func (r *BrokerWatcherReconciler) disableControllers() {
	certControllerState.mutex.Lock()
	defer certControllerState.mutex.Unlock()

	// Stop signer if running
	if r.signer != nil && certControllerState.activeBrokerNamespace != "" {
		r.signer.Stop(certControllerState.activeBrokerNamespace)
	}

	certControllerState.enabled = false
	certControllerState.activeBrokerNamespace = ""
}

func (r *BrokerWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info("Setting up BrokerWatcher - informer on Broker CRs")

	// Resolve broker namespace and start informer
	go func() {

		// If env override is set, use it directly
		if ns := os.Getenv("BROKER_NAMESPACE"); ns != "" {
			r.Log.Info("Using broker namespace from env", "namespace", ns)

			// Enable controllers up front using a direct client
			directClient, _ := client.New(r.Config, client.Options{Scheme: r.Scheme()})
			if directClient != nil {
				if _, err := r.enableControllers(ns, directClient, r.Log.WithValues("source", "env")); err != nil {
					r.Log.Error(err, "Failed to enable controllers from env namespace", "namespace", ns)
				}
			}

			if err := r.startBrokerInformer(ns); err != nil {
				r.Log.Error(err, "Failed to start Broker informer", "namespace", ns)
			}
			return
		}

		// Else, use default broker namespace if not provided via env
		defaultNS := "submariner-k8s-broker"
		r.Log.Info("Using default broker namespace", "namespace", defaultNS)
		if err := r.startBrokerInformer(defaultNS); err != nil {
			r.Log.Error(err, "Failed to start Broker informer", "namespace", defaultNS)
		}
	}()

	// No controller-runtime watches are registered
	return nil
}

// checkExistingBrokers finds existing brokers using direct client (bypasses cache scope)
func (r *BrokerWatcherReconciler) checkExistingBrokers(ctx context.Context) error {
	r.Log.Info("DEBUG: Checking for existing Brokers at startup (bypassing cache)")

	// Create direct client that bypasses cache scope limitations
	directClient, err := client.New(r.Config, client.Options{
		Scheme: r.Scheme(),
	})
	if err != nil {
		return err
	}

	// List all Broker objects cluster-wide
	brokerList := &v1alpha1.BrokerList{}
	if err := directClient.List(ctx, brokerList); err != nil {
		r.Log.Error(err, "Failed to list existing Brokers")
		return err
	}

	if len(brokerList.Items) == 0 {
		r.Log.Info("DEBUG: No existing Broker found at startup")
		return nil
	}

	// Enable controllers for the first broker found
	broker := brokerList.Items[0]
	r.Log.Info("DEBUG: Found existing Broker at startup", "broker", broker.Name, "namespace", broker.Namespace)

	_, err = r.enableControllers(broker.Namespace, directClient, r.Log.WithValues("startup-check"))
	return err
}

// GetCertificateControllerState returns the current state for certificate controllers
func GetCertificateControllerState() (enabled bool, brokerNamespace string) {
	certControllerState.mutex.RLock()
	defer certControllerState.mutex.RUnlock()
	return certControllerState.enabled, certControllerState.activeBrokerNamespace
}

// startBrokerInformer starts a namespaced dynamic informer on Broker CRs
func (r *BrokerWatcherReconciler) startBrokerInformer(namespace string) error {
	r.stopBrokerInformer()

	gvr := schema.GroupVersionResource{Group: "submariner.io", Version: "v1alpha1", Resource: "brokers"}

	dynClient, err := dynamic.NewForConfig(r.Config)
	if err != nil {
		return err
	}

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynClient,
		0,
		namespace,
		nil,
	)

	informer := factory.ForResource(gvr).Informer()

	// Prepare direct client for enable/disable actions
	directClient, _ := client.New(r.Config, client.Options{Scheme: r.Scheme()})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if directClient == nil {
				return
			}
			// On add/update, enable controllers for this namespace
			r.enableControllers(namespace, directClient, r.Log.WithValues("broker-event", "add"))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if directClient == nil {
				return
			}
			r.enableControllers(namespace, directClient, r.Log.WithValues("broker-event", "update"))
		},
		DeleteFunc: func(obj interface{}) {
			// On delete, disable controllers if this was the active namespace
			r.handleBrokerDeleted(namespace)
		},
	})

	stopCh := make(chan struct{})
	r.informerMutex.Lock()
	r.informerStopCh = stopCh
	r.informerActiveNamespace = namespace
	r.informerMutex.Unlock()

	go factory.Start(stopCh)
	go func() { cache.WaitForCacheSync(stopCh, informer.HasSynced) }()

	r.Log.Info("Started Broker CR informer", "namespace", namespace)
	return nil
}

func (r *BrokerWatcherReconciler) stopBrokerInformer() {
	r.informerMutex.Lock()
	defer r.informerMutex.Unlock()

	if r.informerStopCh != nil {
		close(r.informerStopCh)
		r.informerStopCh = nil
		r.Log.Info("Stopped Broker CR informer", "namespace", r.informerActiveNamespace)
	}

	// Also stop the signer when stopping the informer
	if r.signer != nil && r.informerActiveNamespace != "" {
		r.signer.Stop(r.informerActiveNamespace)
		r.Log.Info("Stopped certificate signer", "namespace", r.informerActiveNamespace)
	}

	r.informerActiveNamespace = ""
}

// setupCertificateManagement creates CA certificate and starts CSR signer using Admiral's certificate APIs
func (r *BrokerWatcherReconciler) setupCertificateManagement(namespace string) error {
	// Create dynamic client
	dynClient, err := dynamic.NewForConfig(r.Config)
	if err != nil {
		return err
	}

	// Build REST mapper
	restMapper, err := util.BuildRestMapper(r.Config)
	if err != nil {
		return err
	}

	// Create Admiral signer (only once)
	if r.signer == nil {
		r.signer, err = certificate.NewSigner(certificate.SignerConfig{
			RestConfig: r.Config,
			DynClient:  dynClient,
			RestMapper: restMapper,
		})
		if err != nil {
			return err
		}
	}

	// Start CSR signer for this namespace (CA will be issued automatically)
	if err := r.signer.Start(context.TODO(), namespace); err != nil {
		return err
	}

	r.Log.Info("Certificate management setup completed", "namespace", namespace)
	return nil
}
