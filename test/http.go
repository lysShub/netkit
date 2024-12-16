package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

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

func Gin[T any](t *testing.T, data T) (req *gin.Context) {
	var ch = make(chan struct{})
	g := gin.Default()
	g.POST("/", func(ctx *gin.Context) {
		select {
		case <-ch:
		default:
			req = ctx.Copy()
			close(ch)
		}
	})

	var lis, err = net.Listen("tcp4", ":")
	require.NoError(t, err)
	defer lis.Close()

	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		require.NoError(t, g.RunListener(lis))
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
