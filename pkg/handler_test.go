package pkg

import (
	"context"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

//	func TestSidecarInjector_handleCreate(t *testing.T) {
//		ctx := context.Background()
//		type fields struct {
//			K8sClient                kubernetes.Interface
//			InjectPrefix             string
//			InjectName               string
//			SidecarDataKey           string
//			AllowAnnotationOverrides bool
//			AllowLabelOverrides      bool
//		}
//		type args struct {
//			namespace string
//			pod       v1.Pod
//		}
//		tests := []struct {
//			name      string
//			fields    fields
//			args      args
//			configmap *v1.ConfigMap
//			want      []admission.PatchOperation
//			wantErr   assert.ErrorAssertionFunc
//		}{
//			{
//				name: "pod with no annotations",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod:       v1.Pod{},
//				},
//				want:    nil,
//				wantErr: assert.NoError,
//			},
//			{
//				name: "pod with sidecar annotations no sidecar",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
//						Annotations: map[string]string{
//							"sidecar-injector.expedia.com/inject": "non-sidecar",
//						},
//					}},
//				},
//				configmap: &v1.ConfigMap{},
//				want:      nil,
//				wantErr:   assert.NoError,
//			},
//			{
//				name: "pod with sidecar annotations sidecar with no data",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
//						Annotations: map[string]string{
//							"sidecar-injector.expedia.com/inject": "my-sidecar",
//						},
//					}},
//				},
//				configmap: &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
//					Name: "my-sidecar",
//				}},
//				want:    nil,
//				wantErr: assert.NoError,
//			},
//			{
//				name: "pod with sidecar annotations sidecar with missing sidecar data key",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
//						Annotations: map[string]string{
//							"sidecar-injector.expedia.com/inject": "my-sidecar",
//						},
//					}},
//				},
//				configmap: &v1.ConfigMap{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "my-sidecar",
//					},
//					Data: map[string]string{"wrongKey.yaml": ""},
//				},
//				want:    nil,
//				wantErr: assert.NoError,
//			},
//			{
//				name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
//						Annotations: map[string]string{
//							"sidecar-injector.expedia.com/inject": "my-sidecar",
//						},
//					}},
//				},
//				configmap: &v1.ConfigMap{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "my-sidecar",
//					},
//					Data: map[string]string{"sidecars.yaml": ""},
//				},
//				want:    nil,
//				wantErr: assert.NoError,
//			},
//			{
//				name: "pod with sidecar annotations sidecar with sidecar data key but data empty",
//				fields: fields{
//					K8sClient:      fake.NewSimpleClientset(),
//					InjectPrefix:   "sidecar-injector.expedia.com",
//					InjectName:     "inject",
//					SidecarDataKey: "sidecars.yaml",
//				},
//				args: args{
//					namespace: "test",
//					pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{
//						Annotations: map[string]string{
//							"sidecar-injector.expedia.com/inject": "my-sidecar",
//						},
//					}},
//				},
//				configmap: &v1.ConfigMap{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "my-sidecar",
//					},
//					Data: map[string]string{"sidecars.yaml": `
//	                    - annotations:
//	                        my: annotation
//	                      labels:
//	                        my: label`,
//					},
//				},
//				want: []admission.PatchOperation{
//					{Op: "add", Path: "/metadata/annotations/my", Value: "annotation"},
//					{Op: "add", Path: "/metadata/labels", Value: map[string]string{"my": "label"}}},
//				wantErr: assert.NoError,
//			},
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				patcher := &SidecarInjector{
//					InjectPrefix:             tt.fields.InjectPrefix,
//					InjectName:               tt.fields.InjectName,
//					SidecarDataKey:           tt.fields.SidecarDataKey,
//					AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
//					AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
//				}
//				_, err := patcher.K8sClient.CoreV1().ConfigMaps(tt.args.namespace).Create(ctx, tt.configmap, metav1.CreateOptions{})
//				if err != nil {
//					return
//				}
//				got, err := patcher.handleCreate(ctx, tt.args.namespace, tt.args.pod)
//				if !tt.wantErr(t, err, fmt.Sprintf("handleCreate(%v, %v)", tt.args.namespace, tt.args.pod)) {
//					return
//				}
//				assert.Equalf(t, tt.want, got, "handleCreate(%v, %v)", tt.args.namespace, tt.args.pod)
//			})
//		}
//	}
//
//	func TestSidecarInjector_handleDelete(t *testing.T) {
//		type fields struct {
//			K8sClient                kubernetes.Interface
//			InjectPrefix             string
//			InjectName               string
//			SidecarDataKey           string
//			AllowAnnotationOverrides bool
//			AllowLabelOverrides      bool
//		}
//		type args struct {
//			namespace string
//			pod       v1.Pod
//		}
//		tests := []struct {
//			name    string
//			fields  fields
//			args    args
//			want    []admission.PatchOperation
//			wantErr assert.ErrorAssertionFunc
//		}{
//			{
//				name: "PatchPodDelete is not supported",
//				args: args{
//					namespace: "test",
//					pod:       v1.Pod{},
//				},
//				want:    nil,
//				wantErr: assert.NoError,
//			},
//		}
//		for _, tt := range tests {
//			t.Run(tt.name, func(t *testing.T) {
//				patcher := &SidecarInjector{
//					InjectPrefix:             tt.fields.InjectPrefix,
//					InjectName:               tt.fields.InjectName,
//					SidecarDataKey:           tt.fields.SidecarDataKey,
//					AllowAnnotationOverrides: tt.fields.AllowAnnotationOverrides,
//					AllowLabelOverrides:      tt.fields.AllowLabelOverrides,
//				}
//				ctx := context.Background()
//				got, err := patcher.handleDelete(ctx, tt.args.pod)
//				if !tt.wantErr(t, err, fmt.Sprintf("PatchPodDelete(%v, %v)", tt.args.namespace, tt.args.pod)) {
//					return
//				}
//				assert.Equalf(t, tt.want, got, "PatchPodDelete(%v, %v)", tt.args.namespace, tt.args.pod)
//			})
//		}
//	}
func TestSidecarInjector_handleUpdate(t *testing.T) {
	type fields struct {
		K8sClient                kubernetes.Interface
		InjectPrefix             string
		InjectName               string
		SidecarDataKey           string
		AllowAnnotationOverrides bool
		AllowLabelOverrides      bool
	}
	type args struct {
		request admission.Request
		//oldPod  v1.Pod
		//newPod  v1.Pod
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
				request: admission.Request{
					AdmissionRequest: admissionv1.AdmissionRequest{
						Object:    v1.Pod{},
						OldObject: v1.Pod{},
					},
				},
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
			got := patcher.handleUpdate(ctx, tt.args.oldPod, tt.args.newPod)
			assert.Equalf(t, tt.want, got, "handleUpdate(%v, %v)", tt.args.oldPod, tt.args.newPod)
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
			assert.Equalf(t, tt.want, patcher.sideCarInjectionAnnotation(), "sideCarInjectionAnnotation()")
		})
	}
}
