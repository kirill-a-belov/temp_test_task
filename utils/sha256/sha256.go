package sha256

import (
	"context"
	"crypto/sha256"
	"fmt"

	"playGround/utils/tracing"
)

func Find(ctx context.Context, prefix int64, difficulty int) int64 {
	span, _ := tracing.NewSpan(ctx, "utils.sha256.Find")
	defer span.Close()

	var nonce int64

loop:
	for {
		res := sha256.Sum256([]byte(fmt.Sprint(prefix, nonce)))
		for i := 0; i < difficulty; i++ {
			if res[i] != 0 {
				nonce++
				continue loop
			}
		}

		return nonce
	}
}

func Check(
	ctx context.Context,
	nonce int64,
	prefix int64,
	difficulty int,
) bool {
	span, _ := tracing.NewSpan(ctx, "utils.sha256.Check")
	defer span.Close()

	res := sha256.Sum256([]byte(fmt.Sprint(prefix, nonce)))
	for i := 0; i < difficulty; i++ {
		if res[i] != 0 {
			return false
		}
	}

	return true
}
