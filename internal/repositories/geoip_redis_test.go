package repositories

import (
	"testing"
)

func TestNewRedisDB(t *testing.T) {
	type args struct {
		connstring string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Empty connection string",
			args:    args{connstring: ""},
			wantErr: true,
		},
		{
			name:    "Valid connection string - redis://localhost:6379",
			args:    args{connstring: "redis://localhost:6379"},
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1",
			args:    args{connstring: "redis://user:pass@localhost:6379/1"},
			wantErr: false,
		},
		{
			name:    "Valid connection string - redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2",
			args:    args{connstring: "redis://user:pass@localhost:6379/1?dial_timeout=3&db=1&read_timeout=6s&max_retries=2"},
			wantErr: false,
		},
		{
			name:    "Invalid connection string - redis://user:pass@localhost:6379/db",
			args:    args{connstring: "redis://user:pass@localhost:6379/db"},
			wantErr: true,
		},
		{
			name:    "Invalid connection string - azertyuiop",
			args:    args{connstring: "azertyuiop"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRedisDB(tt.args.connstring)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
