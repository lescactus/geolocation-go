package repositories

import (
	"context"
	"fmt"
	"sync"

	"github.com/lescactus/geolocation-go/models"
)

var (
	ErrInMemoryDBKeyDoesNotExists = &InMemoryDBError{"no value found for key", ""}
)

type InMemoryDBError struct {
	message string
	key     string
}

func (err InMemoryDBError) Error() string {
	return fmt.Sprintf("Error: %s %s", err.message, err.key)
}

type inMemoryDB struct {
	// Hashmap to store the Geo IP infos as key/value
	local map[string]*models.GeoIP
	// Mutex to protect the hashmap from concurrent accesses
	rwm sync.RWMutex
}

// NewInMemoryDB will create a new inMemoryDB.
// It will will instantiate a new hashmap to cache the IP Geolocation information for faster lookups.
func NewInMemoryDB() *inMemoryDB {
	db := &inMemoryDB{}
	db.local = make(map[string]*models.GeoIP)
	return db
}

// Save will add the IP Geolocation info of the given IP address in the hashmap.
// Beause concurrent access to a map isn't safe, Save make use of a sync.RWMutex to lock the map before writing.
func (m *inMemoryDB) Save(ctx context.Context, geoip *models.GeoIP) error {
	m.rwm.Lock()
	defer m.rwm.Unlock()

	m.local[geoip.IP] = geoip

	return nil
}

// Get will retrieve the IP Geolocation info for the given IP address.
// Beause concurrent access to a map isn't safe, Get make use of a sync.RWMutex to lock the map before reading.
// It returns the IP Geolocation info if it's a cache HIT, or an error otherwise.
func (m *inMemoryDB) Get(ctx context.Context, ip string) (*models.GeoIP, error) {
	m.rwm.RLock()
	defer m.rwm.RUnlock()

	v, b := m.local[ip]
	if !b {
		return nil, &InMemoryDBError{
			message: "error: no value found for key",
			key:     ip,
		} //fmt.Errorf("error: no value found for key %s", ip)
	}

	return v, nil
}
