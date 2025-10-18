package utils

import "testing"

func TestGetLocalIPv4(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{name: "ipv4", want: "IP_ADDRESS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIPv4()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalIPv4() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLocalIPv4() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustGetLocalIPv4(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "ipv41", want: "IP_ADDRESS"},
		{name: "ipv42", want: "IP_ADDRESS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustGetLocalIPv4(); got != tt.want {
				t.Errorf("MustGetLocalIPv4() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddLocalIpv4(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "ipv42", args: args{addr: ":8080"}, want: "127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddLocalIpv4(tt.args.addr); got != tt.want {
				t.Errorf("AddLocalIpv4() = %v, want %v", got, tt.want)
			}
		})
	}
}
