package loader

// import (
// 	"testing"
// 	"time"
// )

// func NewTestLiveLoader(args string) (*LiveLoader, error) {
// 	i, _ := time.ParseDuration("1s")
// 	db, err := NewLiveLoader(args, i)
// 	return db, err
// }

// func NewGoodLiveLoader(t testing.TB) *LiveLoader {
// 	db, err := NewTestLiveLoader("-h 127.0.0.1 -u root -proot")
// 	if err != nil {
// 		t.Errorf("Connection error: %s", err)
// 	}
// 	return db
// }

// // NewLiveLoader
// // - should return an error on a bad dsn
// func TestNewLiveLoaderFail(t *testing.T) {
// 	_, err := NewTestLiveLoader("")
// 	if err == nil {
// 		t.Error("No error!")
// 	}
// }

// // - should be able to make a successful connection
// func TestNewLiveLoader(t *testing.T) {
// 	NewGoodLiveLoader(t)
// }

// // Sql Loader implements the Loader interface
// func TestLiveLoaderImplementsLoader(t *testing.T) {
// 	var _ Loader = NewGoodLiveLoader(t)
// }

// // GetSample
// func TestLiveLoaderGetSample(t *testing.T) {
// 	l := NewGoodLiveLoader(t)

// 	samples_ch := l.GetSamples()

// 	sample := <-samples_ch
// 	if sample.Err != nil {
// 		t.Errorf("Sample error: %s", sample.Err)
// 	}

// 	_, err := sample.Status.GetString("uptime")
// 	if err != nil {
// 		t.Error("uptime not in sample")
// 	}

// }

// // GetSamples
// func TestLiveLoaderGetSamples(t *testing.T) {
// 	l := NewGoodLiveLoader(t)
// 	ch := l.GetSamples()

// 	if ch == nil {
// 		t.Error("channel nil")
// 	}
// 	sample := <-ch
// 	_, err := sample.Status.GetString("uptime")
// 	if err != nil {
// 		t.Error("uptime not in sample")
// 	}
// }
