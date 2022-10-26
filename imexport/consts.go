package imexport

import (
	"context"
	"fmt"

	"rxdrag.com/entify/common/contexts"
)

const ARG_SNAPSHOT_ID = "snapshotId"
const ARG_APP_FILE = "appFile"
const ARG_APP_ID = "appId"

const IMPORT_APP = "importApp"
const EXPORT_APP = "exportApp"

const TEMP_DATAS = "temp-datas"
const APP_JON = "app.json"

func getHostPath(ctx context.Context) string {
	return fmt.Sprintf(
		"http://%s/",
		contexts.Values(ctx).Host,
	)
}
