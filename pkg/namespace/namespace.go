package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/seiso/pkg/util"
	"github.com/karrick/tparse/v2"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

const cleanAnnotation = "syn.tools/clean"

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
		checkers      []Checker
	}
	ServiceConfiguration struct {
		Batch bool
	}
	Checker interface {
		NonEmptyNamespaces(context.Context, map[string]struct{}) error
		Name() string
	}
)

// NewNamespacesService creates a new Service instance
func NewNamespacesService(client core.NamespaceInterface, dynamicClient dynamic.Interface, configuration ServiceConfiguration) NamespacesService {
	return NamespacesService{
		client:        client,
		configuration: configuration,
		checkers:      []Checker{NewHelmChecker(), NewResourceChecker(dynamicClient)},
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
	nonEmptyNamespaces := make(map[string]struct{}, len(namespaces))

	for _, checker := range nss.checkers {
		err := checker.NonEmptyNamespaces(ctx, nonEmptyNamespaces)
		if err != nil {
			return nil, fmt.Errorf("could not get %s resources %w", checker.Name(), err)
		}
	}

	for _, ns := range namespaces {
		if _, ok := nonEmptyNamespaces[ns.Name]; ok {
			// Namespace is not empty
			if _, ok := ns.Annotations[cleanAnnotation]; ok && !nss.configuration.Batch {
				log.Warnf("Namespace is annotated for deletion, but not empty. Skipping %q", ns.Name)
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
			log.Infof("Annotated Namespace for deletion: %q", nsCopy.Name)
			if _, err := nss.client.Update(ctx, nsCopy, metav1.UpdateOptions{}); err != nil {
				return nil, err
			}
		}
	}
	return emptyNamespaces, nil
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
			fmt.Println(ns.Name)
		} else {
			log.Infof("Found candidate: %s", ns.Name)
		}
	}
}
