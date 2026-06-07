package logic

import (
	"strconv"

	"google.golang.org/grpc/metadata"
)

func userIDFromMetadata(md metadata.MD) int32 {
	for _, key := range []string{"user_id", "user-id", "x-user-id", "grpc-metadata-user-id"} {
		values := md.Get(key)
		if len(values) == 0 {
			continue
		}
		userID, err := strconv.Atoi(values[0])
		if err == nil && userID > 0 {
			return int32(userID)
		}
	}
	return 0
}
