package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestSidecarInjector_handleCreate(t *testing.T) {
	decoder, err := admission.NewDecoder(scheme)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	type fields struct {
		decoder                  *admission.Decoder
		client                   kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		request admission.Request
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		configmap *corev1.ConfigMap
		want      admission.Response
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "pod with no annotations",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{}, nil),
			},
			configmap: &corev1.ConfigMap{},
			want: admission.Response{
				Patches: []jsonpatch.Operation{},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations no sidecar",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"sidecar-injector.expedia.com/inject": "non-sidecar",
						}}}, nil),
			},
			configmap: &corev1.ConfigMap{},
			want: admission.Response{
				Patches: []jsonpatch.Operation{},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with no data",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar-injector.expedia.com/inject": "my-sidecar",
					},
				}}, nil),
			},
			configmap: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "my-sidecar",
			}},
			want: admission.Response{
				Patches: []jsonpatch.Operation{},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with missing sidecar data key",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"sidecar-injector.expedia.com/inject": "my-sidecar",
						}}}, nil),
			},
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"wrongKey.yaml": ""},
			},
			want: admission.Response{
				Patches: []jsonpatch.Operation{},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"sidecar-injector.expedia.com/inject": "my-sidecar",
						}}}, nil),
			},
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"sidecars.yaml": ""},
			},
			want: admission.Response{
				Patches: []jsonpatch.Operation{},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed:   true,
					PatchType: nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
			fields: fields{
				decoder:        decoder,
				client:         fake.NewSimpleClientset(),
				InjectPrefix:   "sidecar-injector.expedia.com",
				InjectName:     "inject",
				SidecarDataKey: "sidecars.yaml",
			},
			args: args{
				request: mockPodAdmissionsRequest(&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"sidecar-injector.expedia.com/inject": "my-sidecar",
						}}}, nil),
			},
			configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-sidecar",
				},
				Data: map[string]string{"sidecars.yaml": func() string {
					sidecars := []Sidecar{{
						Annotations: map[string]string{
							"my": "annotation",
						},
						Labels: map[string]string{
							"my": "label",
						},
					}}

					if sidecarsStr, err := yaml.Marshal(&sidecars); err != nil {
						panic(err)
					} else {
						return string(sidecarsStr)
					}
				}()},
			},
			want: admission.Response{
				Patches: []jsonpatch.Operation{
					{Operation: "add", Path: "/metadata/labels", Value: map[string]interface{}{"my": "label"}},
					{Operation: "add", Path: "/metadata/annotations/my", Value: "annotation"},
				},
				AdmissionResponse: admissionv1.AdmissionResponse{
					Allowed: true,
					PatchType: func() *admissionv1.PatchType {
						pt := admissionv1.PatchTypeJSONPatch
						return &pt
					}(),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			injector := &SidecarInjector{
				decoder:                  tt.fields.decoder,
				client:                   tt.fields.client,
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			_, err := injector.client.CoreV1().ConfigMaps(tt.args.request.Namespace).Create(ctx, tt.configmap, metav1.CreateOptions{})
			if err != nil {
				return
			}
			got := injector.handleCreate(ctx, tt.args.request)
			if !tt.wantErr(t, err, fmt.Sprintf("handleCreate(%v)", tt.args.request)) {
				return
			}
			assert.Equalf(t, tt.want, got, "handleCreate(%v)", tt.args.request)
		})
	}
}

func TestSidecarInjector_handleDelete(t *testing.T) {
	pod := &corev1.Pod{}

	type fields struct {
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		request admission.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    admission.Response
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "PatchPodDelete is not supported",
			args: args{
				request: mockPodAdmissionsRequest(nil, pod),
			},
			want:    admission.Allowed("delete handled"),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjector{
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			ctx := context.Background()
			got := patcher.handleDelete(ctx, tt.args.request)
			assert.Equalf(t, tt.want, got, "handleDelete(%v)", tt.args.request)
		})
	}
}

func TestSidecarInjector_handleUpdate(t *testing.T) {
	pod := &corev1.Pod{}

	type fields struct {
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		request admission.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    admission.Response
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "handleUpdate is not supported",
			args: args{
				request: mockPodAdmissionsRequest(pod, pod),
			},
			want:    admission.Allowed("update handled"),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patcher := &SidecarInjector{
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			ctx := context.Background()
			got := patcher.handleUpdate(ctx, tt.args.request)
			assert.Equalf(t, tt.want, got, "handleUpdate(%v)", tt.args.request)
		})
	}
}

func TestSidecarInjector_configmapSidecarNames(t *testing.T) {
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
		pod       corev1.Pod
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
				pod:       corev1.Pod{},
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
				pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
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
				pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
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
				pod: corev1.Pod{ObjectMeta: metav1.ObjectMeta{
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
			patcher := &SidecarInjector{
				InjectPrefix:             tt.fields.InjectPrefix,
				InjectName:               tt.fields.InjectName,
				SidecarDataKey:           tt.fields.SidecarDataKey,
				AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
				AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
			}
			got := patcher.configmapSidecarNames(&tt.args.pod)
			assert.Equalf(t, tt.want, got, "configmapSidecarNames(%v, %v)", tt.args.namespace, tt.args.pod)
		})
	}
}

func TestSidecarInjector_sideCarInjectionAnnotation(t *testing.T) {
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
			patcher := &SidecarInjector{
				InjectPrefix: tt.fields.InjectPrefix,
				InjectName:   tt.fields.InjectName,
			}
			assert.Equalf(t, tt.want, patcher.sidecarInjectionAnnotation(), "sidecarInjectionAnnotation()")
		})
	}
}

func mockPodAdmissionsRequest(pod *corev1.Pod, oldPod *corev1.Pod) admission.Request {
	podRaw, err := json.Marshal(pod)
	if err != nil {
		panic(err)
	}
	oldPodRaw, err := json.Marshal(oldPod)
	if err != nil {
		panic(err)
	}
	admissionRequest := admissionv1.AdmissionRequest{
		UID:       "my-uid",
		Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		Name:      "my-pod",
		Namespace: "default",
	}
	if pod != nil {
		admissionRequest.Object = runtime.RawExtension{
			Raw: podRaw,
		}
		admissionRequest.Operation = "CREATE"
		admissionRequest.Namespace = pod.Namespace
	}
	if oldPod != nil {
		admissionRequest.OldObject = runtime.RawExtension{
			Raw: oldPodRaw,
		}
		admissionRequest.Operation = "DELETE"
		admissionRequest.Namespace = oldPod.Namespace
	}
	if pod != nil && oldPod != nil {
		admissionRequest.Operation = "UPDATE"
	}
	return admission.Request{
		AdmissionRequest: admissionRequest,
	}
}
