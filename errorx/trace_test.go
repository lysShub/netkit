package errorx_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/lysShub/netkit/errorx"
	"github.com/tidwall/gjson"
)

func e0() error {
	err := errors.New("xxx")
	return err
}

func e1() error {
	err := e0()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func t0() string {
	buff := bytes.NewBuffer(nil)
	l := slog.New(slog.NewJSONHandler(buff, nil))

	err := e1()
	l.Error(err.Error(), errorx.Trace(err))
	return buff.String()
}

func Test_Trace(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		out := t0()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go:16", out)

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.String, res2.Type)
		require.Contains(t, res2.Str, "trace_test.go:21", out)

		res3 := gjson.Get(out, "trace.2")
		require.Equal(t, gjson.String, res3.Type)
		require.Contains(t, res3.Str, "trace_test.go:33", out)

		res4 := gjson.Get(out, "trace.3")
		require.Equal(t, gjson.String, res4.Type)
		require.Contains(t, res4.Str, "trace_test.go:34", out)

		res5 := gjson.Get(out, "trace.4")
		require.Equal(t, gjson.String, res5.Type)
		require.Contains(t, res5.Str, "trace_test.go:41", out)

		res6 := gjson.Get(out, "trace.5")
		require.Equal(t, gjson.Null, res6.Type)
	})

	t.Run("nil", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		l.Error(" ", errorx.Trace(nil))
		out := b.String()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go:71", out)

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.Null, res2.Type)
	})

	t.Run("global err", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		a := errorx.Trace(globalErr)
		require.Equal(t, slog.KindLogValuer, a.Value.Kind())

		l.Error(" ", a)
		out := b.String()
		require.Contains(t, out, "trace_test.go:86")
	})
}

var globalErr = errors.New("global error")
