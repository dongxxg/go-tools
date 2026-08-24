package grpc

import (
	"context"
	"strings"

	"github.com/dongxxg/go-tools/common"
	"google.golang.org/grpc/metadata"
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