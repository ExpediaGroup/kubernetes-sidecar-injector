package webhook

import (
	"github.com/expediagroup/kubernetes-sidecar-injector/pkg/admission"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_createAnnotationPatches(t *testing.T) {
	type args struct {
		newMap      map[string]string
		existingMap map[string]string
		override    bool
	}
	tests := []struct {
		name string
		args args
		want []admission.PatchOperation
	}{
		{
			name: "blah",
			args: args{
				newMap:      map[string]string{"my": "annotation"},
				existingMap: map[string]string{},
			},
			want: []admission.PatchOperation{{
				Op:    "add",
				Path:  "/metadata/annotations",
				Value: map[string]string{"my": "annotation"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createAnnotationPatches(tt.args.newMap, tt.args.existingMap, tt.args.override); !reflect.DeepEqual(got, tt.want) {
				assert.Fail(t, "annotation patching failed", "createAnnotationPatches() = %v, want %v", got, tt.want)
			}
		})
	}
}
