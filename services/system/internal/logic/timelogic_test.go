package logic

import (
	"context"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/testkit/assertx"
	system "github.com/falconfan123/Go-mall/services/system/pb"
	"github.com/stretchr/testify/require"
)

func TestTimeLogic(t *testing.T) {
	t.Parallel()

	now := time.Now().UnixMilli()
	logic := NewTimeLogic(context.Background(), nil)

	resp, err := logic.Time(&system.TimeReq{})
	require.NoError(t, err)
	assertx.RequireWithinRange(t, resp.Now, now, time.Now().UnixMilli())
}
