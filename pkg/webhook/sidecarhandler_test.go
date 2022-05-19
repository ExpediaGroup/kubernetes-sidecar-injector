package webhook

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

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
