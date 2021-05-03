package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/seiso/pkg/util"
	"github.com/karrick/tparse"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	cleanAnnotation  = "syn.tools/clean"
	helmDriverSecret = "secret"
)

var (
	resources = []schema.GroupVersionResource{
		{Version: "v1", Resource: "pods"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "extensions", Version: "v1beta1", Resource: "daemonsets"},
		{Group: "extensions", Version: "v1beta1", Resource: "deployments"},
	}
)

type (
	NamespacesService struct {
		configuration ServiceConfiguration
		client        core.NamespaceInterface
		dynamicClient dynamic.Interface
	}
	ServiceConfiguration struct {
		Batch bool
	}
)

// NewNamespacesService creates a new Service instance
func NewNamespacesService(client core.NamespaceInterface, dynamicClient dynamic.Interface, configuration ServiceConfiguration) NamespacesService {
	return NamespacesService{
		client:        client,
		dynamicClient: dynamicClient,
		configuration: configuration,
	}
}

func (nss NamespacesService) List(ctx context.Context, listOptions metav1.ListOptions) ([]corev1.Namespace, error) {
	ns, err := nss.client.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return ns.Items, nil
}

func (nss NamespacesService) GetEmptyFor(ctx context.Context, namespaces []corev1.Namespace, duration string) ([]corev1.Namespace, error) {
	now := time.Now()
	emptyNamespaces := []corev1.Namespace{}
	namespaceMap := make(map[string]struct{}, len(namespaces))

	if err := nss.getHelmReleases(ctx, namespaces, namespaceMap); err != nil {
		return nil, fmt.Errorf("could not get Helm releases %w", err)
	}
	if err := nss.getResources(ctx, namespaces, namespaceMap); err != nil {
		return nil, fmt.Errorf("could not get Resources %w", err)
	}

	for _, ns := range namespaces {
		if _, ok := namespaceMap[ns.Name]; ok {
			// Namespace is not empty
			if _, ok := ns.Annotations[cleanAnnotation]; ok && !nss.configuration.Batch {
				log.Warnf("Non empty namespace is annotated for deletion, skip: %q", ns.Name)
			}
			continue
		}

		ts, ok := ns.Annotations[cleanAnnotation]

		if ok {
			emptySince, err := time.Parse(util.TimeFormat, ts)
			if err != nil {
				return nil, err
			}
			deleteAt, err := tparse.AddDuration(emptySince, duration)
			if err != nil {
				return nil, err
			}
			if now.After(deleteAt) {
				emptyNamespaces = append(emptyNamespaces, ns)
			}
		} else {
			nsCopy := ns.DeepCopy()
			if nsCopy.Annotations == nil {
				nsCopy.Annotations = make(map[string]string, 1)
			}
			nsCopy.Annotations[cleanAnnotation] = now.UTC().Format(util.TimeFormat)
			log.Infof("Annotated namespace for deletion: %q", nsCopy.Name)
			if _, err := nss.client.Update(ctx, nsCopy, metav1.UpdateOptions{}); err != nil {
				return nil, err
			}
		}
	}
	return emptyNamespaces, nil
}

func (nss NamespacesService) getHelmReleases(ctx context.Context, namespaces []corev1.Namespace, namespaceMap map[string]struct{}) error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(genericclioptions.NewConfigFlags(true), "", helmDriverSecret, func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format, v))
	}); err != nil {
		return err
	}

	listAction := action.NewList(actionConfig)
	listAction.AllNamespaces = true
	releases, err := listAction.Run()
	if err != nil {
		return err
	}
	for _, release := range releases {
		if release.Info.Deleted.IsZero() {
			// Found an active Release in this namespace
			namespaceMap[release.Namespace] = struct{}{}
		}
	}
	return nil
}

func (nss NamespacesService) getResources(ctx context.Context, namespaces []corev1.Namespace, namespaceMap map[string]struct{}) error {
	for _, r := range resources {
		resourceList, err := nss.dynamicClient.Resource(r).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil
		}

		for _, resource := range resourceList.Items {
			if resource.GetDeletionTimestamp().IsZero() {
				// Found active resource in namespace
				namespaceMap[resource.GetNamespace()] = struct{}{}
			}
		}
	}
	return nil
}

func (nss NamespacesService) Delete(ctx context.Context, namespaces []corev1.Namespace) error {
	for _, ns := range namespaces {
		err := nss.client.Delete(ctx, ns.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if nss.configuration.Batch {
			fmt.Println(ns.Name)
		} else {
			log.Infof("Deleted Namespace %q", ns.Name)
		}
	}
	return nil
}

func (nss NamespacesService) Print(namespaces []corev1.Namespace) {
	if len(namespaces) == 0 {
		log.Info("Nothing found to be deleted.")
	}

	for _, ns := range namespaces {
		if nss.configuration.Batch {
			fmt.Println(ns.GetName())
		} else {
			log.Infof("Found candidate: %s", ns.Name)
		}
	}
}
