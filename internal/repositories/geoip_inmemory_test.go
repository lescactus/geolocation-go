package repositories

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/lescactus/geolocation-go/internal/models"
)

func TestNewInMemoryDB(t *testing.T) {
	tests := []struct {
		name string
		want *inMemoryDB
	}{
		{
			name: "New inMemoryDB",
			want: &inMemoryDB{local: make(map[string]*models.GeoIP)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInMemoryDB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInMemoryDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDBSave(t *testing.T) {
	type fields struct {
		local map[string]*models.GeoIP
		rwm   sync.RWMutex
	}
	type args struct {
		ctx   context.Context
		geoip *models.GeoIP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Insert 1.1.1.1",
			fields: fields{local: make(map[string]*models.GeoIP)},
			args: args{
				ctx: context.Background(),
				geoip: &models.GeoIP{
					IP:          "1.1.1.1.",
					CountryCode: "AU",
					CountryName: "Australia",
					City:        "Sydney",
					Latitude:    -27.4766,
					Longitude:   -153.0166,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &inMemoryDB{
				local: tt.fields.local,
				rwm:   tt.fields.rwm,
			}
			if err := m.Save(tt.args.ctx, tt.args.geoip); (err != nil) != tt.wantErr {
				t.Errorf("inMemoryDB.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInMemoryDBGet(t *testing.T) {
	type fields struct {
		local map[string]*models.GeoIP
		rwm   sync.RWMutex
	}
	type args struct {
		ctx context.Context
		ip  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.GeoIP
		wantErr bool
	}{
		{
			name: "Get existing IP",
			fields: fields{local: map[string]*models.GeoIP{"1.1.1.1": {
				IP:          "1.1.1.1.",
				CountryCode: "AU",
				CountryName: "Australia",
				City:        "Sydney",
				Latitude:    -27.4766,
				Longitude:   -153.0166,
			}}},
			args: args{ctx: context.Background(), ip: "1.1.1.1"},
			want: &models.GeoIP{
				IP:          "1.1.1.1.",
				CountryCode: "AU",
				CountryName: "Australia",
				City:        "Sydney",
				Latitude:    -27.4766,
				Longitude:   -153.0166,
			},
			wantErr: false,
		},
		{
			name: "Get non existing IP",
			fields: fields{local: map[string]*models.GeoIP{"1.1.1.1": {
				IP:          "1.1.1.1.",
				CountryCode: "AU",
				CountryName: "Australia",
				City:        "Sydney",
				Latitude:    -27.4766,
				Longitude:   -153.0166,
			}}},
			args:    args{ctx: context.Background(), ip: "2.2.2.2"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &inMemoryDB{
				local: tt.fields.local,
				rwm:   tt.fields.rwm,
			}
			got, err := m.Get(tt.args.ctx, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("inMemoryDB.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inMemoryDB.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInMemoryDBStatus(t *testing.T) {
	m := NewInMemoryDB()
	var wg sync.WaitGroup
	ch := make(chan error, 1)

	wg.Add(1)
	go m.Status(context.Background(), &wg, ch)
	wg.Wait()

	err := <-ch

	if err != nil {
		t.Errorf("inMemoryDB.Get() error = %v, wantErr %v", err, false)
		return
	}
}

func BenchmarkInMemoryDBGet_EntryInMemoryDB(b *testing.B) {
	m := NewInMemoryDB()
	m.Save(context.Background(), &models.GeoIP{IP: "1.1.1.1"})

	b.ResetTimer()
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		m.Get(ctx, "1.1.1.1")
	}
}

func BenchmarkInMemoryDBGet_EntryNotInMemoryDB(b *testing.B) {
	m := NewInMemoryDB()
	m.Save(context.Background(), &models.GeoIP{IP: "1.1.1.1"})

	b.ResetTimer()
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		m.Get(ctx, "2.2.2.2")
	}
}

func BenchmarkInMemoryDBSave_EntryInMemoryDB(b *testing.B) {
	m := NewInMemoryDB()
	m.Save(context.Background(), &models.GeoIP{IP: "1.1.1.1"})

	b.ResetTimer()
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		m.Save(ctx, &models.GeoIP{IP: "1.1.1.1"})
	}
}

func BenchmarkInMemoryDBSave_EntryNotInMemoryDB(b *testing.B) {
	m := NewInMemoryDB()
	m.Save(context.Background(), &models.GeoIP{IP: "1.1.1.1"})

	b.ResetTimer()
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		m.Save(ctx, &models.GeoIP{IP: "2.2.2.2"})
	}
}
