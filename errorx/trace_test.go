package errorx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/lysShub/netkit/errorx"
	"github.com/stretchr/testify/require"
)

func Test_Frames(t *testing.T) {

	t.Run("empty", func(t *testing.T) {
		var f errorx.Frames

		rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		rec.Add(slog.Attr{Key: "trace", Value: slog.AnyValue(f)})

		var b = &bytes.Buffer{}
		slog.NewJSONHandler(b, nil).Handle(context.Background(), rec)

		require.NotContains(t, "trace", b.String())

		lit, err := f.JsonLiteral(nil)
		require.NoError(t, err)
		require.Nil(t, lit)
	})

	t.Run("JsonLiteral", func(t *testing.T) {
		e := e0()

		var f = errorx.Trace(e).Value.Any().(errorx.Frames)

		lit, err := f.JsonLiteral(nil)
		require.NoError(t, err)

		// lit = bytes.ReplaceAll(lit, []byte("trace_test.go"), []byte(`trace_"test.go`))

		// not always
		require.True(t, json.Valid(lit))

		// {
		// 	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
		// 	rec.Add(errorx.Trace(e))
		// 	var b = &bytes.Buffer{}
		// 	slog.NewJSONHandler(b, nil).Handle(context.Background(), rec)
		// 	str := b.String()
		// 	fmt.Println()
		// 	fmt.Println(str)
		// 	fmt.Println()
		// 	fmt.Println(string(lit))
		// 	fmt.Println()
		// }
	})
}
