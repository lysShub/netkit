package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestXxxxx(t *testing.T) {

	type Data struct {
		Name string
	}

	var data = Data{
		Name: "sdfasdf",
	}

	req := Post(t, data)
	fmt.Println(req)

	w := &httptest.ResponseRecorder{}

	ctx := gin.CreateTestContextOnly(w, gin.Default())
	ctx.Request = req

	var d Data
	require.NoError(t, ctx.Bind(&d))

	ctx.JSON(http.StatusOK, gin.H{"msg": "ok"})

}
