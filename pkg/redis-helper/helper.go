package rh

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func GenRandomKeys(prefix string, num int) []string {
	var result []string
	for i := 0; i < num; i++ {
		result = append(result, fmt.Sprintf("%s%s", prefix, uuid.New().String()))
	}

	return result
}

func BatchSetKeys(ctx context.Context, keys []string, ttl int, cli redis.UniversalClient) error {
	pip := cli.Pipeline()
	for _, k := range keys {
		pip.Set(ctx, k, "", time.Duration(ttl)*time.Second)
	}

	_, err := pip.Exec(ctx)

	return err
}

func BatchCheckNotExistKeys(ctx context.Context, keys []string, cli redis.UniversalClient) ([]string, error) {
	pip := cli.Pipeline()

	var (
		notExistKey []string
		queryResult = map[string]*redis.IntCmd{}
	)

	for _, key := range keys {
		queryResult[key] = pip.Exists(ctx, key)
	}

	_, err := pip.Exec(ctx)
	if err != nil {
		return nil, err
	}

	for key, result := range queryResult {
		if result.Val() == 0 {
			notExistKey = append(notExistKey, key)
		}
	}

	return notExistKey, nil
}

func LoopCheckKeysAllExist(timeoutCtx context.Context, keysToCheck []string, cli redis.UniversalClient) ([]string, error) {
	var (
		notExistKeys []string
		err          error
	)

	notExistKeys, err = BatchCheckNotExistKeys(timeoutCtx, keysToCheck, cli)

	if err != nil {
		return nil, err
	}

	if len(notExistKeys) == 0 {
		return nil, nil
	}

	keysToCheck = notExistKeys

	for {
		select {
		case <-timeoutCtx.Done():
			return notExistKeys, err
		case <-time.After(time.Second):
			notExistKeys, err = BatchCheckNotExistKeys(timeoutCtx, keysToCheck, cli)
			if err != nil {
				return nil, err
			}

			if len(notExistKeys) == 0 {
				return nil, nil
			}

			keysToCheck = notExistKeys
		}
	}
}
