package bot

const (
	START_JOINING = "join (%d channels): start"
	DONE_JOINING  = "join (%d/%d): done"
	FAILED_JOIN   = "join (%d/%d): failed (%s)"
)
const (
	START_FORWARDING  = "forward (%d): start"
	DONE_FORWARDING   = "forward (%d): done"
	FAILED_FORWARDING = "forward (%d): failed (%s)"
)

const (
	DONE_ARCHIVE   = "archive: done"
	FAILED_ARCHIVE = "archive: failed (%s)"
)

const DONE_ALL = "all done"
