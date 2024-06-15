package bot

import "github.com/celestix/gotgproto/types"

func filterReqJoin(m *types.Message) bool {
	return m.Text == "/j"
}
func filterReqFwd(m *types.Message) bool {
	return m.Text == "/f"
}
