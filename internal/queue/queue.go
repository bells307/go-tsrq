package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Очередь элементов Redis, основанная на временных отметках
type RedisTimestampRoundQueue struct {
	// Непосредственно клиент Redis
	client *redis.Client
	// Имя очереди
	name string
	// Периодичность изъятия из очереди
	dequeuePeriod time.Duration
	// Время жизни элемента в очереди
	ttl time.Duration
}

// Результат операции изъятия из очереди
type DequeueResult struct {
	Id   string
	Data []byte
}

func NewRedisTimestampRoundQueue(client *redis.Client, name string, dequeuePeriod, ttl time.Duration) *RedisTimestampRoundQueue {
	return &RedisTimestampRoundQueue{client, name, dequeuePeriod, ttl}
}

// Запуск периодической очистки очереди
func (q *RedisTimestampRoundQueue) RunCleaner(ctx context.Context, cleanPeriod time.Duration) {
	go func() {
		ticker := time.NewTicker(cleanPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				q.cleanup(ctx)
			}
		}
	}()
}

// Добавить элемент в очередь
func (q *RedisTimestampRoundQueue) Enqueue(ctx context.Context, id string, data any) error {
	// Элемент нужно добавить сразу в 2 индекса (zset'а) - по времени создания и по времени последнего
	// изъятия из очереди. Также мы должны добавить полезные данные в hashmap.
	// Так как операция должна быть атомарной, мы используем транзакцию.
	_, err := q.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		// Добавляем в индекс по времени создания
		z := redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: id,
		}

		pipe.ZAddNX(ctx, q.createTimestampIndexName(), z)

		// Добавляем в индекс по времени последнего изъятия
		z = redis.Z{
			// Устанавливаем score = 0, чтобы поднять приоритет
			Score:  0,
			Member: id,
		}

		pipe.ZAddNX(ctx, q.lastDequeueIndexName(), z)

		// Добавляем в хэшмап
		pipe.HSetNX(ctx, q.dataHashmapName(), id, data)

		return nil
	})

	return err
}

// Берет следующий элемент из очереди
func (q *RedisTimestampRoundQueue) Dequeue(ctx context.Context) (*DequeueResult, error) {
	deqResult := (*DequeueResult)(nil)

	tranFn := func(tx *redis.Tx) error {
		z := redis.ZRangeArgs{
			Key:     q.lastDequeueIndexName(),
			Start:   "-inf",
			Stop:    "+inf",
			ByScore: true,
			ByLex:   false,
			Rev:     false,
			Offset:  0,
			Count:   1,
		}

		// Получаем id элемента на выдачу
		zRangeRes, err := tx.ZRangeArgsWithScores(ctx, z).Result()
		if err != nil {
			return fmt.Errorf("error while ZRANGE command executing: %v", err)
		}

		if len(zRangeRes) == 0 {
			// Данных нет
			return nil
		}

		id := zRangeRes[0].Member.(string)
		lastDeqTs := zRangeRes[0].Score

		if lastDeqTs > float64(time.Now().Unix()-int64(q.dequeuePeriod.Seconds())) {
			// Прошло слишком мало времени с момента последнего изъятия из очереди
			return nil
		}

		// Достаем данные из хэшмапа по id
		hGetRes, err := tx.HGet(ctx, q.dataHashmapName(), string(id)).Result()
		if err != nil {
			return fmt.Errorf("error while HGET command executing: %v", err)
		}

		deqResult = &DequeueResult{id, []byte(hGetRes)}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			// Обновляем время в индексе по времени последнего изъятия из очереди
			z := redis.Z{
				Score:  float64(time.Now().Unix()),
				Member: id,
			}

			pipe.ZAdd(ctx, q.lastDequeueIndexName(), z)
			return nil
		})

		return err
	}

	err := TransactionLoop(ctx, q.client, nil, tranFn, q.lastDequeueIndexName(), q.createTimestampIndexName(), q.dataHashmapName())

	return deqResult, err
}

// Удалить элемент из очереди
func (q *RedisTimestampRoundQueue) Remove(ctx context.Context, id string) error {
	_, err := q.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.ZRem(ctx, q.createTimestampIndexName(), id)
		pipe.ZRem(ctx, q.lastDequeueIndexName(), id)
		pipe.HDel(ctx, q.dataHashmapName(), id)

		return nil
	})

	return err
}

// Проверка элемента на существование в очереди
func (q *RedisTimestampRoundQueue) Exists(ctx context.Context, id string) (bool, error) {
	_, err := q.client.ZScore(ctx, q.createTimestampIndexName(), id).Result()

	if err == nil {
		return true, nil
	} else if err == redis.Nil {
		return false, nil
	} else {
		return false, err
	}
}

// Подсчет количества элементов
func (q *RedisTimestampRoundQueue) Count(ctx context.Context) (int64, error) {
	return q.client.ZCard(ctx, q.createTimestampIndexName()).Result()
}

// Очистка устаревших элементов
func (q *RedisTimestampRoundQueue) cleanup(ctx context.Context) error {
	opt := redis.ZRangeBy{
		Min: "1",
		Max: fmt.Sprint(time.Now().Unix() - int64(q.ttl.Seconds())),
	}

	// Получем идентификаторы элементов на удаление
	ids, _ := q.client.ZRangeByScore(ctx, q.createTimestampIndexName(), &opt).Result()

	_, err := q.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.ZRem(ctx, q.createTimestampIndexName(), ids)
		pipe.ZRem(ctx, q.lastDequeueIndexName(), ids)
		pipe.HDel(ctx, q.dataHashmapName(), ids...)
		return nil
	})

	return err
}

// Имя индекса (zset'а) элементов по времени последнего изъятия из очереди
func (q *RedisTimestampRoundQueue) lastDequeueIndexName() string {
	return fmt.Sprintf("%s_ld", q.name)
}

// Имя индекса (zset'а) элементов по времени добавления в очередь
func (q *RedisTimestampRoundQueue) createTimestampIndexName() string {
	return fmt.Sprintf("%s_ct", q.name)
}

// Имя hashmap'а, в котором хранятся полезные данные
func (q *RedisTimestampRoundQueue) dataHashmapName() string {
	return fmt.Sprintf("%s_data", q.name)
}
