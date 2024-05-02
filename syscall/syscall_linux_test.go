//go:build linux
// +build linux

package syscall_test

import (
	"testing"
)

func Test_IoctlTSO(t *testing.T) {
	// ifi := "lo"
	// var on = func() bool {
	// 	msg, err := exec.Command("ethtool", "-k", ifi).CombinedOutput()
	// 	require.NoError(t, err)
	// 	rows := strings.Split(string(msg), "\n")
	// 	for _, e := range rows {
	// 		if strings.Contains(e, "tcp-segmentation-offload") {
	// 			return strings.Contains(e, " on")
	// 		}
	// 	}
	// 	panic("")
	// }

	// init := on()
	// defer func() { helper.IoctlTSO(ifi, !init) }()

	// err := helper.IoctlTSO(ifi, !init)
	// require.NoError(t, err)
	// o := on()
	// fmt.Println(o)
	// require.Equal(t, !init, o)
}
