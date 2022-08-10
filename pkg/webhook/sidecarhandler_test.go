package webhook

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSidecarInjectorPatcher_PatchPodCreate(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		K8sClient                kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		namespace string
		pod       v1.Pod
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		configmap *v1.ConfigMap
		want      []admission.PatchOperation
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "pod with no annotations",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod:       v1.Pod{},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations no sidecar",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "non-sidecar",
					},
				}},
			},
			configmap: &v1.ConfigMap{},
			want:      nil,
			wantErr:   assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with no data",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}},
			},
			configmap: &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "my-sidecar",
			}},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with missing sidecar data key",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}},
			},
			configmap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"wrongKey.yaml": ""},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}},
			},
			configmap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"sidecars.yaml": ""},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
			fields: fields{
				K8sClient:      fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}},
			},
			configmap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"sidecars.yaml": `
                     - annotations:
                         my: annotation
                       labels:
                         my: label`,
				},
			},
			want: []admission.PatchOperation{
				{Op: "add", Path: "/metadata/annotations/my", Value: "annotation"},
				{Op: "add", Path: "/metadata/labels", Value: map[string]string{"my": "label"}}},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjectorPatcher{
				K8sClient:                tt.fields.K8sClient,
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			_, err := patcher.K8sClient.CoreV1().ConfigMaps(tt.args.namespace).Create(ctx, tt.configmap, metav1.CreateOptions{})
			if err != nil {
				return
			}
			got, err := patcher.PatchPodCreate(ctx, tt.args.namespace, tt.args.pod)
			if !tt.wantErr(t, err, fmt.Sprintf("PatchPodCreate(%v, %v)", tt.args.namespace, tt.args.pod)) {
				return
			}
			assert.Equalf(t, tt.want, got, "PatchPodCreate(%v, %v)", tt.args.namespace, tt.args.pod)
		})
	}
}

func TestSidecarInjectorPatcher_PatchPodDelete(t *testing.T) {
	type fields struct {
		K8sClient                kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		namespace string
		pod       v1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []admission.PatchOperation
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "PatchPodDelete is not supported",
			args: args{
				namespace: "test",
				pod:       v1.Pod{},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjectorPatcher{
				K8sClient:                tt.fields.K8sClient,
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			ctx := context.Background()
			got, err := patcher.PatchPodDelete(ctx, tt.args.namespace, tt.args.pod)
			if !tt.wantErr(t, err, fmt.Sprintf("PatchPodDelete(%v, %v)", tt.args.namespace, tt.args.pod)) {
				return
			}
			assert.Equalf(t, tt.want, got, "PatchPodDelete(%v, %v)", tt.args.namespace, tt.args.pod)
		})
	}
}

func TestSidecarInjectorPatcher_PatchPodUpdate(t *testing.T) {
	type fields struct {
		K8sClient                kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		namespace string
		oldPod    v1.Pod
		newPod    v1.Pod
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []admission.PatchOperation
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "PatchPodUpdate is not supported",
			args: args{
				namespace: "test",
				oldPod:    v1.Pod{},
				newPod:    v1.Pod{},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjectorPatcher{
				K8sClient:                tt.fields.K8sClient,
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			ctx := context.Background()
			got, err := patcher.PatchPodUpdate(ctx, tt.args.namespace, tt.args.oldPod, tt.args.newPod)
			if !tt.wantErr(t, err, fmt.Sprintf("PatchPodUpdate(%v, %v, %v)", tt.args.namespace, tt.args.oldPod, tt.args.newPod)) {
				return
			}
			assert.Equalf(t, tt.want, got, "PatchPodUpdate(%v, %v, %v)", tt.args.namespace, tt.args.oldPod, tt.args.newPod)
		})
	}
}

func TestSidecarInjectorPatcher_configmapSidecarNames(t *testing.T) {
	type fields struct {
		K8sClient                kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		namespace string
		pod       v1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "configmap sidecars has no annotations",
			args: args{
				namespace: "test",
				pod:       v1.Pod{},
			},
			fields: fields{
				K8sClient:    fake.NewSimpleClientset(),
				InjectPrefix: "sidecar-injector.expedia.com",
				InjectName:   "inject",
			},
			want: nil,
		},
		{
			name: "configmap sidecars has annotations but no sidecars",
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"test": "annotation",
					},
				}},
			},
			fields: fields{
				K8sClient:    fake.NewSimpleClientset(),
				InjectPrefix: "sidecar-injector.expedia.com",
				InjectName:   "inject",
			},
			want: nil,
		},
		{
			name: "configmap sidecars has a sidecar",
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}},
			},
			fields: fields{
				K8sClient:    fake.NewSimpleClientset(),
				InjectPrefix: "sidecar-injector.expedia.com",
				InjectName:   "inject",
			},
			want: []string{"my-sidecar"},
		},
		{
			name: "configmap sidecars has multiple sidecar",
			args: args{
				namespace: "test",
				pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar,my-sidecar2",
					},
				}},
			},
			fields: fields{
				K8sClient:    fake.NewSimpleClientset(),
				InjectPrefix: "sidecar-injector.expedia.com",
				InjectName:   "inject",
			},
			want: []string{"my-sidecar", "my-sidecar2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjectorPatcher{
				K8sClient:                tt.fields.K8sClient,
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			got := patcher.configmapSidecarNames(tt.args.namespace, tt.args.pod)
			assert.Equalf(t, tt.want, got, "configmapSidecarNames(%v, %v)", tt.args.namespace, tt.args.pod)
		})
	}
}

func TestSidecarInjectorPatcher_sideCarInjectionAnnotation(t *testing.T) {
	type fields struct {
		K8sClient    kubernetes.Interface
		InjectPrefix string
		InjectName   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "sidecar injection annotation properly constructed",
			fields: fields{
				K8sClient:    fake.NewSimpleClientset(),
				InjectPrefix: "sidecar-injector.expedia.com",
				InjectName:   "inject",
			},
			want: "sidecar-injector.expedia.com/inject",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjectorPatcher{
				K8sClient:    tt.fields.K8sClient,
				InjectPrefix: tt.fields.InjectPrefix,
				InjectName:   tt.fields.InjectName,
			}
			assert.Equalf(t, tt.want, patcher.sideCarInjectionAnnotation(), "sideCarInjectionAnnotation()")
		})
	}
}

func Test_createArrayPatches(t *testing.T) {
	type args[T any] struct {
		newCollection      []T
		existingCollection []T
		path               string
	}
	containerTests := []struct {
		name string
		args args[v1.Container]
		want []admission.PatchOperation
	}{
		{
			name: "test patching initContainer first",
			args: args[v1.Container]{
				newCollection:      []v1.Container{{Name: "Test"}},
				existingCollection: []v1.Container{},
				path:               "/spec/initContainers",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/spec/initContainers",
				Value: []v1.Container{{Name: "Test"}},
			}},
		}, {
			name: "test patching initContainer not first",
			args: args[v1.Container]{
				newCollection:      []v1.Container{{Name: "Test2"}},
				existingCollection: []v1.Container{{Name: "Test"}},
				path:               "/spec/initContainers",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/spec/initContainers/-",
				Value: v1.Container{Name: "Test2"},
			}},
		}, {
			name: "test patching multiple initContainer not first",
			args: args[v1.Container]{
				newCollection:      []v1.Container{{Name: "Test2"}, {Name: "Test3"}},
				existingCollection: []v1.Container{{Name: "Test"}},
				path:               "/spec/initContainers",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/spec/initContainers/-",
				Value: v1.Container{Name: "Test2"},
			}, {
				Op:    "add",
				Path:  "/spec/initContainers/-",
				Value: v1.Container{Name: "Test3"},
			}},
		},
	}
	for _, tt := range containerTests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, createArrayPatches(tt.args.newCollection, tt.args.existingCollection, tt.args.path), "createArrayPatches(%v, %v, %v)", tt.args.newCollection, tt.args.existingCollection, tt.args.path)
		})
	}
}

func Test_createObjectPatches(t *testing.T) {
	type args struct {
		newMap      map[string]string
		existingMap map[string]string
		path        string
		override    bool
	}
	tests := []struct {
		name string
		args args
		want []admission.PatchOperation
	}{
		{
			name: "test patching empty annotation",
			args: args{
				newMap:      map[string]string{"my": "annotation"},
				existingMap: map[string]string{},
				path:        "/metadata/annotations",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/annotations/my",
				Value: "annotation",
			}},
		},
		{
			name: "test patching empty annotation with forward slash",
			args: args{
				newMap:      map[string]string{"example.com/my": "annotation"},
				existingMap: map[string]string{},
				path:        "/metadata/annotations",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/annotations/example.com~1my",
				Value: "annotation",
			}},
		},
		{
			name: "test patching empty annotation with tilde",
			args: args{
				newMap:      map[string]string{"example.com~my": "annotation"},
				existingMap: map[string]string{},
				path:        "/metadata/annotations",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/annotations/example.com~0my",
				Value: "annotation",
			}},
		},
		{
			name: "test patching nil annotation",
			args: args{
				newMap:      map[string]string{"my": "annotation"},
				existingMap: nil,
				path:        "/metadata/annotations",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/annotations",
				Value: map[string]string{"my": "annotation"},
			}},
		},
		{
			name: "test patching annotation no override",
			args: args{
				newMap:      map[string]string{"my": "annotation"},
				existingMap: map[string]string{"my": "override-annotation"},
				path:        "/metadata/annotations",
			},
			want: nil,
		},
		{
			name: "test patching annotation with override",
			args: args{
				newMap:      map[string]string{"my": "annotation"},
				existingMap: map[string]string{"my": "override-annotation"},
				path:        "/metadata/annotations",
				override:    true,
			},
			want: []admission.PatchOperation{{
				Op:    "replace",
				Path:  "/metadata/annotations/my",
				Value: "annotation",
			}},
		},
		{
			name: "test patching empty labels",
			args: args{
				newMap:      map[string]string{"my": "label"},
				existingMap: map[string]string{},
				path:        "/metadata/labels",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/labels/my",
				Value: "label",
			}},
		},
		{
			name: "test patching empty labels with forward slash",
			args: args{
				newMap:      map[string]string{"example.com/my": "label"},
				existingMap: map[string]string{},
				path:        "/metadata/labels",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/labels/example.com~1my",
				Value: "label",
			}},
		},
		{
			name: "test patching empty labels with tilde",
			args: args{
				newMap:      map[string]string{"example.com~my": "label"},
				existingMap: map[string]string{},
				path:        "/metadata/labels",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/labels/example.com~0my",
				Value: "label",
			}},
		},
		{
			name: "test patching nil labels",
			args: args{
				newMap:      map[string]string{"my": "label"},
				existingMap: nil,
				path:        "/metadata/labels",
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/labels",
				Value: map[string]string{"my": "label"},
			}},
		},
		{
			name: "test patching label no override",
			args: args{
				newMap:      map[string]string{"my": "label"},
				existingMap: map[string]string{"my": "override-label"},
				path:        "/metadata/labels",
			},
			want: nil,
		},
		{
			name: "test patching label with override",
			args: args{
				newMap:      map[string]string{"my": "label"},
				existingMap: map[string]string{"my": "override-label"},
				path:        "/metadata/labels",
				override:    true,
			},
			want: []admission.PatchOperation{{
				Op:    "replace",
				Path:  "/metadata/labels/my",
				Value: "label",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createObjectPatches(tt.args.newMap, tt.args.existingMap, tt.args.path, tt.args.override); !reflect.DeepEqual(got, tt.want) {
				assert.Fail(t, "annotation patching failed", "createObjectPatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
