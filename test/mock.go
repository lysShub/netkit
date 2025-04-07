package test

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/stretchr/testify/require"
)

type mockTest struct{}

func T() *mockTest { return &mockTest{} }

var _ require.TestingT = (*mockTest)(nil)

func (m *mockTest) Errorf(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	fmt.Fprintln(os.Stderr)

	os.Exit(1)
}

func (m *mockTest) FailNow() {
	fmt.Fprintln(os.Stderr, "Fail")
	os.Exit(1)
}

type printTest struct {
	failed atomic.Bool
}

func P() *printTest { return &printTest{} }

func (m *printTest) Errorf(format string, args ...interface{}) {
	m.failed.CompareAndSwap(false, true)
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
}

func (m *printTest) FailNow() {
	m.failed.CompareAndSwap(false, true)
	fmt.Fprintln(os.Stderr, "FailNow")
}

func (m *printTest) Failed() bool {
	return m.failed.Load()
}
