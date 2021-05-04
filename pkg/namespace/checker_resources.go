package namespace

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

const resourceCheckerName = "Resources"

type ResourceChecker struct {
	dynamicClient dynamic.Interface
}

func NewResourceChecker(dynamicClient dynamic.Interface) *ResourceChecker {
	return &ResourceChecker{dynamicClient: dynamicClient}
}

func (rc ResourceChecker) Name() string {
	return resourceCheckerName
}

func (rc ResourceChecker) NonEmptyNamespaces(ctx context.Context, namespaceMap map[string]struct{}) error {
	for _, r := range resources {
		resourceList, err := rc.dynamicClient.Resource(r).List(ctx, metav1.ListOptions{})
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
