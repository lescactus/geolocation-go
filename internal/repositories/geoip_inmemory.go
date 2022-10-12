package repositories

import (
	"context"
	"fmt"
	"sync"

	"github.com/lescactus/geolocation-go/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ErrInMemoryDBKeyDoesNotExists = &InMemoryDBError{"no value found for key", ""}
)

// Prometheus metrics
var (
	inMemoryItemSaved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "in_memory_items_saved_total",
		Help: "The total number of saved items in the in-memory database",
	})
	inMemoryItemRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "in_memory_items_read_total",
		Help: "The total number of read items from the in-memory database",
	})
	inMemoryItemFailedRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "in_memory_items_failed_read_total",
		Help: "The total number of failed read items from the in-memory database",
	})
)

type InMemoryDBError struct {
	message string
	key     string
}

func (err InMemoryDBError) Error() string {
	return fmt.Sprintf("error: %s %s", err.message, err.key)
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

	// Increment Prometheus counter
	inMemoryItemSaved.Inc()

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
		// Increment Prometheus counter
		inMemoryItemFailedRead.Inc()
		return nil, &InMemoryDBError{
			message: "no value found for key",
			key:     ip,
		}
	}

	// Increment Prometheus counter
	inMemoryItemRead.Inc()

	return v, nil
}

// Status will retrieve the status of the inMemoryDB.
func (m *inMemoryDB) Status(ctx context.Context, wg *sync.WaitGroup, ch chan error) {
	defer wg.Done()

	_, _ = m.Get(ctx, "")

	ch <- nil
}
