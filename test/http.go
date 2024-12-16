package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// Post construct http post request
func Post[T any](t *testing.T, data T) (req *http.Request) {
	var ch = make(chan struct{})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-ch:
		default:
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			req = r.Clone(context.Background())
			req.Body = io.NopCloser(bytes.NewBuffer(b))
			close(ch)
		}
	})

	var lis, err = net.Listen("tcp4", ":")
	require.NoError(t, err)
	defer lis.Close()

	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		require.NoError(t, http.Serve(lis, http.DefaultServeMux))
		return nil
	})
	eg.Go(func() error {
		time.Sleep(time.Second)
		var b = bytes.NewBuffer(nil)
		require.NoError(t, json.NewEncoder(b).Encode(data))
		url := fmt.Sprintf("http://%s", lis.Addr())
		resp, err := http.Post(url, "application/json", b)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		return nil
	})
	<-ch
	require.NoError(t, lis.Close())
	require.NoError(t, eg.Wait())

	return req
}

type Response[T any] struct {
	*httptest.ResponseRecorder
}

func NewResp[T any]() *Response[T] {
	return &Response[T]{
		ResponseRecorder: &httptest.ResponseRecorder{
			Body: bytes.NewBuffer(nil),
		},
	}
}

func (r *Response[T]) Unmarshal() (T, error) {
	var val T
	err := json.NewDecoder(r.Body).Decode(&val)
	return val, err
}

// Gin construct gin context
// Example:
//
//	ctx,resp := Gin[RegistReq,RegistResp](t, RegistReq{Name:"Tom"})
//	ctr.Regist(ctx)
//	require.Equal(t, http.StatusOK, resp.Code) // or other check
func Gin[Req, Resp any](t *testing.T, data Req) (*gin.Context, *Response[Resp]) {
	w := NewResp[Resp]()

	ctx := gin.CreateTestContextOnly(w, gin.Default())
	ctx.Request = Post(t, data)
	return ctx, w
}
