package cfg

import (
	"time"
)

//KubernetesResource kubernetes resource interface
type KubernetesResource struct {
	name              string
	namespace         string
	kind              string
	creationTimestamp time.Time
	labels            map[string]string
}

//GetName returns name of the resource
func (res *KubernetesResource) GetName() string { return res.name }

//SetName sets name of the resource
func (res *KubernetesResource) SetName(name string) { res.name = name }

//GetNamespace returns namespace of the resource
func (res *KubernetesResource) GetNamespace() string { return res.namespace }

//SetNamespace sets namespace of the resource
func (res *KubernetesResource) SetNamespace(namespace string) { res.namespace = namespace }

//GetKind returns type of the resource
func (res *KubernetesResource) GetKind() string { return res.kind }

//SetKind sets type of the resource
func (res *KubernetesResource) SetKind(kind string) { res.kind = kind }

//GetCreationTimestamp returns creation date of the resource
func (res *KubernetesResource) GetCreationTimestamp() time.Time { return res.creationTimestamp }

//SetCreationTimestamp sets creation date of the resource
func (res *KubernetesResource) SetCreationTimestamp(creationTimestamp time.Time) {
	res.creationTimestamp = creationTimestamp
}

//GetLabels returns labels of the resource
func (res *KubernetesResource) GetLabels() map[string]string { return res.labels }

//SetLabels sets labels of the resource
func (res *KubernetesResource) SetLabels(labels map[string]string) { res.labels = labels }

//NewResource resource constructor
func NewResource(name, namespace, kind string, creationTimestamp time.Time, labels map[string]string) KubernetesResource {
	return KubernetesResource{
		name:              name,
		namespace:         namespace,
		kind:              kind,
		creationTimestamp: creationTimestamp,
		labels:            labels,
	}
}

//NewConfigMapResource config map resource constructor
func NewConfigMapResource(name, namespace string, creationTimestamp time.Time, labels map[string]string) KubernetesResource {
	return NewResource(name, namespace, "ConfigMap", creationTimestamp, labels)
}

//NewSecretResource secret resource constructor
func NewSecretResource(name, namespace string, creationTimestamp time.Time, labels map[string]string) KubernetesResource {
	return NewResource(name, namespace, "Secret", creationTimestamp, labels)
}
