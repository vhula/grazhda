package pkg

import "testing"

func TestValidateNameOrAll(t *testing.T) {
	tests := []struct {
		name    string
		all     bool
		wantErr bool
	}{
		{name: "", all: false, wantErr: true},
		{name: "", all: true, wantErr: false},
		{name: "foo", all: false, wantErr: false},
		{name: "foo", all: true, wantErr: true},
	}
	for _, tt := range tests {
		label := "name=" + tt.name + ",all="
		if tt.all {
			label += "true"
		} else {
			label += "false"
		}
		t.Run(label, func(t *testing.T) {
			err := validateNameOrAll(tt.name, tt.all)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNameOrAll(%q, %v) error = %v, wantErr %v",
					tt.name, tt.all, err, tt.wantErr)
			}
		})
	}
}
