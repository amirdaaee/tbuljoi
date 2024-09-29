package bot

import (
	"strings"

	"github.com/celestix/gotgproto/types"
)

func isFromSelf(m *types.Message) bool {
	return m.Out
}
func filterReqJoin(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/j"
}
func filterReqUnjoin(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/uj"
}
func filterReqFwd(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/f"
}
func filterReqFwdArch(m *types.Message) bool {
	return isFromSelf(m) && m.Text == "/ff"
}
func filterReqAFSet(m *types.Message) bool {
	return isFromSelf(m) && strings.HasPrefix(m.Text, "/af ")
}
