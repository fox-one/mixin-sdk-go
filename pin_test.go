package mixin

import (
	"testing"
)

func Test_ValidatePinPattern(t *testing.T) {
	type args struct {
		pin string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid pin",
			args: args{
				pin: "123456",
			},
			wantErr: false,
		},
		{
			name: "empty pin",
			args: args{
				pin: "",
			},
			wantErr: true,
		},
		{
			name: "short pin",
			args: args{
				pin: "123",
			},
			wantErr: true,
		},
		{
			name: "long pin",
			args: args{
				pin: "12345678",
			},
			wantErr: true,
		},
		{
			name: "pin with non-numeric",
			args: args{
				pin: "123 23",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePinPattern(tt.args.pin); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePinPattern() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
