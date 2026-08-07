package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
	"unitechs.com/unios-dice/uni-base/core/common"
)

func ExtractRequestID(ctx context.Context) string {
	requestID := ""
	md, _ := metadata.FromIncomingContext(ctx)
	reqIds, ok := md[strings.ToLower(common.HeaderRequestID)]
	if ok {
		requestID = reqIds[0]
	}
	return requestID
}
