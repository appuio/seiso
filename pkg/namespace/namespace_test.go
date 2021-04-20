package namespace

import (
	"context"
	"testing"
	"time"

	"github.com/appuio/seiso/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func Test_GetEmptyFor(t *testing.T) {

	tests := map[string]struct {
		objs        []runtime.Object
		deleteAfter string
		want        []string
		wantErr     bool
	}{
		"NoNamespaces": {
			objs:    []runtime.Object{},
			want:    []string{},
			wantErr: false,
		},
		"Delete2Namespaces": {
			objs: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-1",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(-1 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-2",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(-24 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
			},
			deleteAfter: "1s",
			want:        []string{"test-ns-1", "test-ns-2"},
			wantErr:     false,
		},
		"DeleteNoNamespaces": {
			objs: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-1",
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-2",
						Annotations: map[string]string{
							"test": "some",
						},
					},
				},
			},
			deleteAfter: "24h",
			want:        []string{},
			wantErr:     false,
		},
		"DeleteNotYetNamespaces": {
			objs: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-1",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Format(util.TimeFormat),
						},
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-2",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(2 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-ns-3",
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "delete-test-ns",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(-23 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
			},
			deleteAfter: "24h",
			want:        []string{},
			wantErr:     false,
		},
		"InvalidAnnotationNamespace": {
			objs: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "invalid",
						Annotations: map[string]string{
							cleanAnnotation: "this-is-invalid",
						},
					},
				},
			},
			want:    []string{},
			wantErr: true,
		},
		"NamespaceNotEmpty": {
			objs: []runtime.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ns1",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(-48 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "ns1",
					},
				},
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ns2",
						Annotations: map[string]string{
							cleanAnnotation: time.Now().UTC().Add(-48 * time.Hour).Format(util.TimeFormat),
						},
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "ns2",
					},
				},
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			logrus.SetLevel(logrus.DebugLevel)
			ctx := context.Background()
			clientset := fake.NewSimpleClientset(tt.objs...)
			fakeClient := clientset.CoreV1().Namespaces()
			fakeDynamicClient := dynFake.NewSimpleDynamicClient(scheme.Scheme, tt.objs...)

			service := NewNamespacesService(fakeClient, fakeDynamicClient, ServiceConfiguration{})

			allNamespaces, err := service.List(ctx, metav1.ListOptions{})
			if !tt.wantErr {
				assert.NoError(t, err)
			}

			list, err := service.GetEmptyFor(ctx, allNamespaces, tt.deleteAfter)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			nsNames := []string{}
			for _, ns := range list {
				nsNames = append(nsNames, ns.Name)
			}
			assert.ElementsMatch(t, tt.want, nsNames)
		})
	}
}
