package errorx_test

import (
	"bytes"
	"context"
	stderr "errors"
	"log/slog"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/lysShub/netkit/errorx"
	"github.com/tidwall/gjson"
)

func Test_Loggeer(t *testing.T) {
	{
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		err := c()
		l.LogAttrs(context.TODO(), slog.LevelError, err.Error(), errorx.TraceAttr(err))
		out := b.String()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go")

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.String, res2.Type)
		require.Contains(t, res2.Str, "trace_test.go")

		res3 := gjson.Get(out, "trace.2")
		require.Equal(t, gjson.String, res3.Type)
		require.Contains(t, res3.Str, "trace_test.go")
	}

	{
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		err := c()
		l.Error(err.Error(), errorx.TraceAttr(err))
		out := b.String()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go")

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.String, res2.Type)
		require.Contains(t, res2.Str, "trace_test.go")

		res3 := gjson.Get(out, "trace.2")
		require.Equal(t, gjson.String, res3.Type)
		require.Contains(t, res3.Str, "trace_test.go")
	}

	{
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		err := stderr.New("xxx")
		l.Error(err.Error(), errorx.TraceAttr(err))
		out := b.String()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go")

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.Null, res2.Type)
	}

	{
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		var err error = nil
		l.Error("nil", errorx.TraceAttr(err))
		out := b.String()

		res1 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_test.go")

		res2 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.Null, res2.Type)
	}
}

func c() error {
	e := b()
	if e != nil {
		err := errors.WithStack(errors.New("c-fail"))

		return err
	}

	return nil
}

func b() error {
	e := a()
	if e != nil {
		return errors.WithStack(e)
	}

	return nil
}

func a() error {
	e := errors.New("xxx")
	return errors.WithStack(e)
}
