package util

import "testing"

func TestUrlEncode(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				uri: "http://localhost:8989/send?name=wiken&age=18",
			},
			want:    "http://localhost:8989/send?age=18&name=wiken",
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				uri: "http://localhost:8989/send?name=花样百出&age=18",
			},
			want:    "http://localhost:8989/send?age=18&name=%E8%8A%B1%E6%A0%B7%E7%99%BE%E5%87%BA",
			wantErr: false,
		},
		{
			name: "1",
			args: args{
				uri: "http://localhost:8989/send?name=GE100/0/0/1&age=18",
			},
			want:    "http://localhost:8989/send?age=18&name=GE100%2F0%2F0%2F1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UrlEncode(tt.args.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("UrlEncode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UrlEncode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
