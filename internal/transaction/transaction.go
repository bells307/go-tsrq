package transaction

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

// Запуск транзакции в цикле
func TransactionLoop(ctx context.Context, client *redis.Client, maxRetries *int, tranFn func(*redis.Tx) error, keys ...string) error {
	// Счетчик попыток
	retriesCount := 0

	for {
		if maxRetries != nil {
			if retriesCount >= *maxRetries {
				return errors.New("reached maximum number of retries")
			}
		}

		err := client.Watch(ctx, tranFn, keys...)
		if err == nil {
			// Ошибки нет, можно выходить из цикла
			return nil
		}

		if err == redis.TxFailedErr {
			// Ключ изменился в ходе транзакции, пробуем еще
			retriesCount++
			continue
		}

		// Остальные ошибки прокидываем вверх
		return err
	}
}
