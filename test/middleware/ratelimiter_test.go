package middleware_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"

	"github.com/parxyws/cozybox/internal/middleware"
)

func redisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "2" {
		return addr
	}
	return "127.0.0.1:6379"
}

func newTestRedis(t *testing.T) *redisclient.Client {
	t.Helper()
	rdb := redisclient.NewClient(&redisclient.Options{Addr: redisAddr()})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available at %s: %v", redisAddr(), err)
	}
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func newRateLimitedRouter(rdb *redisclient.Client, cfg middleware.RateLimitConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.NewRateLimiter(rdb, cfg))
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func doRequest(t *testing.T, r *gin.Engine, ip string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", ip)
	req.RemoteAddr = ip + ":12345"
	r.ServeHTTP(w, req)
	return w
}

func TestRateLimiter_WithinLimit(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:within:%d", os.Getpid())

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 3,
		Window:      mustParseDuration("10s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	})

	for i := 0; i < 3; i++ {
		w := doRequest(t, r, "10.0.0.1")
		if w.Code != 200 {
			t.Fatalf("req %d: expected 200, got %d. headers: %v", i+1, w.Code, w.Header())
		}
		assertHeader(t, w, "X-RateLimit-Limit", "3")
		assertHeader(t, w, "X-RateLimit-Remaining", strconv.Itoa(2-i))
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:exceed:%d", os.Getpid())

	cfg := middleware.RateLimitConfig{
		MaxRequests: 2,
		Window:      mustParseDuration("10s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	}
	r := newRateLimitedRouter(rdb, cfg)

	for i := 0; i < 2; i++ {
		w := doRequest(t, r, "10.0.0.2")
		if w.Code != 200 {
			t.Fatalf("req %d: expected 200, got %d", i+1, w.Code)
		}
	}

	w := doRequest(t, r, "10.0.0.2")
	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "too many requests" {
		t.Errorf("unexpected error: %v", body["error"])
	}
	if _, ok := body["retry_after"]; !ok {
		t.Error("missing retry_after in body")
	}
	assertHeaderPresent(t, w, "Retry-After")
}

func TestRateLimiter_ResetAfterHeaders(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:headers:%d", os.Getpid())

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 1,
		Window:      mustParseDuration("5s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	})

	w := doRequest(t, r, "10.0.0.3")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	assertHeaderPresent(t, w, "X-RateLimit-Limit")
	assertHeader(t, w, "X-RateLimit-Remaining", "0")

	reset := w.Header().Get("X-RateLimit-Reset")
	resetVal, err := strconv.Atoi(reset)
	if err != nil || resetVal <= 0 || resetVal > 5 {
		t.Errorf("X-RateLimit-Reset should be 1-5, got %q", reset)
	}

	w = doRequest(t, r, "10.0.0.3")
	if w.Code != 429 {
		t.Fatalf("expected 429, got %d", w.Code)
	}
	assertHeader(t, w, "X-RateLimit-Remaining", "0")

	retry := w.Header().Get("Retry-After")
	retryVal, err := strconv.Atoi(retry)
	if err != nil || retryVal <= 0 || retryVal > 5 {
		t.Errorf("Retry-After should be 1-5, got %q", retry)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:ips:%d", os.Getpid())

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 1,
		Window:      mustParseDuration("10s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	})

	w1 := doRequest(t, r, "10.0.0.10")
	if w1.Code != 200 {
		t.Fatalf("ip1 req1: expected 200, got %d", w1.Code)
	}

	w2 := doRequest(t, r, "10.0.0.11")
	if w2.Code != 200 {
		t.Fatalf("ip2 req1: expected 200, got %d (different IPs should not affect each other)", w2.Code)
	}
}

func TestRateLimiter_FailOpenOnRedisDown(t *testing.T) {
	rdb := redisclient.NewClient(&redisclient.Options{
		Addr:         "127.0.0.1:16379",
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		MaxRetries:   0,
	})
	defer rdb.Close()

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 1,
		Window:      mustParseDuration("10s"),
		KeyFunc:     middleware.RateLimitByIP("test:failopen"),
	})

	for i := 0; i < 10; i++ {
		w := doRequest(t, r, "10.0.0.99")
		if w.Code != 200 {
			t.Fatalf("req %d: expected 200 on Redis down, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_ConcurrentRequests(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:concurrent:%d", os.Getpid())

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 10,
		Window:      mustParseDuration("5s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	})

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes, rejected := 0, 0

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := doRequest(t, r, "10.0.0.20")
			mu.Lock()
			if w.Code == 200 {
				successes++
			} else if w.Code == 429 {
				rejected++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	if successes != 10 {
		t.Errorf("expected 10 successes, got %d", successes)
	}
	if rejected != 10 {
		t.Errorf("expected 10 rejected, got %d", rejected)
	}
}

func TestRateLimiter_RedisKeyStoredCorrectly(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:stored:%d", os.Getpid())
	ip := "10.0.0.50"

	cfg := middleware.RateLimitConfig{
		MaxRequests: 5,
		Window:      mustParseDuration("30s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	}
	r := newRateLimitedRouter(rdb, cfg)

	expectedKey := fmt.Sprintf("ratelimit:%s:%s", prefix, ip)

	for i := 1; i <= 3; i++ {
		w := doRequest(t, r, ip)
		if w.Code != 200 {
			t.Fatalf("req %d: expected 200, got %d", i, w.Code)
		}

		val, err := rdb.Get(context.Background(), expectedKey).Result()
		if err != nil {
			t.Fatalf("req %d: redis key %q not found: %v", i, expectedKey, err)
		}
		if val != strconv.Itoa(i) {
			t.Errorf("req %d: expected counter %d, got %s", i, i, val)
		}
	}

	ttl, err := rdb.TTL(context.Background(), expectedKey).Result()
	if err != nil {
		t.Fatalf("failed to get TTL for %q: %v", expectedKey, err)
	}
	if ttl <= 0 {
		t.Errorf("expected TTL > 0, got %v (key should have expiry set)", ttl)
	}
	if ttl > 30*time.Second {
		t.Errorf("expected TTL <= 30s, got %v", ttl)
	}

	t.Logf("key %q counter=3 ttl=%v — verify manually: redis-cli GET %q && redis-cli TTL %q", expectedKey, ttl, expectedKey, expectedKey)
}

func TestRateLimiter_KeyExpiresAfterWindow(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := fmt.Sprintf("test:expire:%d", os.Getpid())
	ip := "10.0.0.60"
	window := 2 * time.Second

	r := newRateLimitedRouter(rdb, middleware.RateLimitConfig{
		MaxRequests: 3,
		Window:      window,
		KeyFunc:     middleware.RateLimitByIP(prefix),
	})

	w := doRequest(t, r, ip)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	key := fmt.Sprintf("ratelimit:%s:%s", prefix, ip)
	exists, _ := rdb.Exists(context.Background(), key).Result()
	if exists != 1 {
		t.Fatalf("key %q should exist after first request", key)
	}

	time.Sleep(window + 100*time.Millisecond)

	exists, _ = rdb.Exists(context.Background(), key).Result()
	if exists != 0 {
		t.Errorf("key %q should have expired after %v", key, window)
	}
}

func TestRateLimiter_KeyNamePattern(t *testing.T) {
	rdb := newTestRedis(t)
	prefix := "test:pattern"
	ip := "192.168.1.100"

	cfg := middleware.RateLimitConfig{
		MaxRequests: 1,
		Window:      mustParseDuration("10s"),
		KeyFunc:     middleware.RateLimitByIP(prefix),
	}
	r := newRateLimitedRouter(rdb, cfg)

	doRequest(t, r, ip)

	expectedKey := fmt.Sprintf("ratelimit:%s:%s", prefix, ip)
	keys, _ := rdb.Keys(context.Background(), "ratelimit:test:pattern:*").Result()
	found := false
	for _, k := range keys {
		if k == expectedKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("key %q not found in Redis (keys: %v)", expectedKey, keys)
	}

	val, _ := rdb.Get(context.Background(), expectedKey).Result()
	if val != "1" {
		t.Errorf("expected counter 1, got %q", val)
	}

	t.Logf("key %q value=%q — verify manually: redis-cli GET \"%s\" && redis-cli GETRANGE \"%s\" 0 -1", expectedKey, val, expectedKey, expectedKey)
}

func assertHeader(t *testing.T, w *httptest.ResponseRecorder, key, expected string) {
	t.Helper()
	got := w.Header().Get(key)
	if got != expected {
		t.Errorf("header %s: expected %q, got %q", key, expected, got)
	}
}

func assertHeaderPresent(t *testing.T, w *httptest.ResponseRecorder, key string) {
	t.Helper()
	if w.Header().Get(key) == "" {
		t.Errorf("header %s should be present", key)
	}
}

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}
