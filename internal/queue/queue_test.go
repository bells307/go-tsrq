package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
)

func TestEnqueueDequeue(t *testing.T) {
	ctx := context.Background()

	client := testClient(ctx)

	q := NewRedisTimestampRoundQueue(client, "tsrtest", time.Second*1, time.Second*10)

	err := q.Enqueue(ctx, "123", "somedata")
	if err != nil {
		t.Error(err)
	}

	testDequeue(ctx, q, t, &DequeueResult{Id: "123", Data: []byte("somedata")})

	// Ждем маленький промежуток времени, ожидаем, что элемента не будет
	time.Sleep(time.Millisecond * 10)

	testDequeue(ctx, q, t, nil)

	// Ждем еще, ожидаем, что элемент появится
	time.Sleep(time.Second * 1)

	testDequeue(ctx, q, t, &DequeueResult{Id: "123", Data: []byte("somedata")})
}

func TestMultiEnqueue(t *testing.T) {
	ctx := context.Background()

	client := testClient(ctx)

	q := NewRedisTimestampRoundQueue(client, "tsrtest", time.Second*1, time.Second*10)

	testEnqueue(ctx, q, t, "123", "somedata")
	testEnqueue(ctx, q, t, "456", "somedata2")
	testEnqueue(ctx, q, t, "789", "somedata3")

	testDequeue(ctx, q, t, &DequeueResult{Id: "123", Data: []byte("somedata")})
	testDequeue(ctx, q, t, &DequeueResult{Id: "456", Data: []byte("somedata2")})
	testDequeue(ctx, q, t, &DequeueResult{Id: "789", Data: []byte("somedata3")})
	testDequeue(ctx, q, t, nil)

	time.Sleep(time.Second * 1)

	testDequeue(ctx, q, t, &DequeueResult{Id: "123", Data: []byte("somedata")})
	testDequeue(ctx, q, t, &DequeueResult{Id: "456", Data: []byte("somedata2")})
	testDequeue(ctx, q, t, &DequeueResult{Id: "789", Data: []byte("somedata3")})
	testDequeue(ctx, q, t, nil)
}

func TestRemove(t *testing.T) {
	ctx := context.Background()

	client := testClient(ctx)

	q := NewRedisTimestampRoundQueue(client, "tsrtest", time.Second*1, time.Second*10)

	testEnqueue(ctx, q, t, "123", "somedata")

	err := q.Remove(ctx, "123")
	if err != nil {
		t.Error(err)
	}

	testDequeue(ctx, q, t, nil)

}

func TestCleanup(t *testing.T) {
	ctx := context.Background()
	client := testClient(ctx)
	q := NewRedisTimestampRoundQueue(client, "tsrtest", time.Second*1, time.Second*1)
	q.RunCleaner(ctx, time.Second*1)

	testEnqueue(ctx, q, t, "123", "somedata")
	testEnqueue(ctx, q, t, "456", "somedata2")
	testEnqueue(ctx, q, t, "789", "somedata3")

	count, err := q.Count(ctx)
	if err != nil {
		t.Error(err)
	}

	if count != 3 {
		t.Errorf("expected 3 elements, got %v", count)
	}

	time.Sleep(time.Second * 2)

	count, err = q.Count(ctx)
	if err != nil {
		t.Error(err)
	}

	if count != 0 {
		t.Errorf("expected 0 elements, got %v", count)
	}
}

func testClient(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// TODO: dont flush
	client.FlushDB(ctx)

	return client
}

func testEnqueue(ctx context.Context, q *RedisTimestampRoundQueue, t *testing.T, id string, data any) {
	err := q.Enqueue(ctx, id, data)
	if err != nil {
		t.Error(err)
	}
}

func testDequeue(ctx context.Context, q *RedisTimestampRoundQueue, t *testing.T, exp *DequeueResult) {
	deqRes, err := q.Dequeue(ctx)
	if err != nil {
		t.Error(err)
	}

	if exp == nil {
		if deqRes != nil {
			t.Errorf("expected nil dequeue result, got %v", deqRes)
		}
		return
	}

	if !cmp.Equal(*deqRes, *exp) {
		t.Errorf("expected %v dequeue result, got %v", deqRes, exp)
	}
}
