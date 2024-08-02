package errorx_test

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	_ "encoding/csv"
	_ "encoding/gob"
	_ "encoding/hex"
	_ "encoding/json"
	_ "encoding/pem"
	_ "encoding/xml"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/lysShub/netkit/errorx"
	"github.com/tidwall/gjson"
)

func e0() error {
	err := errors.New("xxx") // line 24
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

		res0 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res0.Type)
		require.Contains(t, res0.Str, "trace_pc_test.go:24", out)

		res1 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "trace_pc_test.go:29", out)

		res2 := gjson.Get(out, "trace.2")
		require.Equal(t, gjson.String, res2.Type)
		require.Contains(t, res2.Str, "trace_pc_test.go:41", out)

		res3 := gjson.Get(out, "trace.3")
		require.Equal(t, gjson.String, res3.Type)
		require.Contains(t, res3.Str, "trace_pc_test.go:42", out)

		res4 := gjson.Get(out, "trace.4")
		require.Equal(t, gjson.String, res4.Type)
		require.Contains(t, res4.Str, "trace_pc_test.go:49", out)

		res5 := gjson.Get(out, "trace.5")
		require.Equal(t, gjson.String, res5.Type)
		require.Contains(t, res5.Str, "testing.go", out) // if test: testing.go:1175 , if main: runtime/asm_amd64.s:1695, if go: not this frame

		res6 := gjson.Get(out, "trace.6")
		require.Equal(t, gjson.Null, res6.Type, out)
	})

	t.Run("nil", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		l.Error(" ", errorx.Trace(nil))
		out := b.String()

		res0 := gjson.Get(out, "trace.0")
		require.Equal(t, gjson.String, res0.Type)
		require.Contains(t, res0.Str, "trace_pc_test.go:83", out)

		res1 := gjson.Get(out, "trace.1")
		require.Equal(t, gjson.String, res1.Type)
		require.Contains(t, res1.Str, "testing.go", out)

		res2 := gjson.Get(out, "trace.2")
		require.Equal(t, gjson.Null, res2.Type, out)
	})

	t.Run("global err", func(t *testing.T) {
		b := bytes.NewBuffer(nil)
		l := slog.New(slog.NewJSONHandler(b, nil))

		a := errorx.Trace(globalErr)
		require.Equal(t, slog.KindLogValuer, a.Value.Kind())

		l.Error(" ", a)
		out := b.String()
		require.Contains(t, out, "trace_pc_test.go:102", out) // it not work perfect
	})
}

var globalErr = errors.New("global error") // not recommend

func Test_Trace_Go(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		go func() {
			out := t0()

			res0 := gjson.Get(out, "trace.0")
			require.Equal(t, gjson.String, res0.Type)
			require.Contains(t, res0.Str, "trace_pc_test.go:24", out)

			res1 := gjson.Get(out, "trace.1")
			require.Equal(t, gjson.String, res1.Type)
			require.Contains(t, res1.Str, "trace_pc_test.go:29", out)

			res2 := gjson.Get(out, "trace.2")
			require.Equal(t, gjson.String, res2.Type)
			require.Contains(t, res2.Str, "trace_pc_test.go:41", out)

			res3 := gjson.Get(out, "trace.3")
			require.Equal(t, gjson.String, res3.Type)
			require.Contains(t, res3.Str, "trace_pc_test.go:42", out)

			res4 := gjson.Get(out, "trace.4")
			require.Equal(t, gjson.String, res4.Type)
			require.Contains(t, res4.Str, "trace_pc_test.go:116", out)

			res5 := gjson.Get(out, "trace.5")
			require.Equal(t, gjson.Null, res5.Type, out)
		}()
		time.Sleep(time.Second)
	})

	t.Run("nil", func(t *testing.T) {
		go func() {
			b := bytes.NewBuffer(nil)
			l := slog.New(slog.NewJSONHandler(b, nil))

			l.Error(" ", errorx.Trace(nil))
			out := b.String()

			res0 := gjson.Get(out, "trace.0")
			require.Equal(t, gjson.String, res0.Type)
			require.Contains(t, res0.Str, "trace_pc_test.go:149", out)

			res1 := gjson.Get(out, "trace.1")
			require.Equal(t, gjson.Null, res1.Type, out)
		}()
		time.Sleep(time.Second)
	})

	t.Run("global err", func(t *testing.T) {
		go func() {
			b := bytes.NewBuffer(nil)
			l := slog.New(slog.NewJSONHandler(b, nil))

			a := errorx.Trace(globalErr)
			require.Equal(t, slog.KindLogValuer, a.Value.Kind())

			l.Error(" ", a)
			out := b.String()
			require.Contains(t, out, "trace_pc_test.go:167", out)
		}()
		time.Sleep(time.Second)
	})
}
