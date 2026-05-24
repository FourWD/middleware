package kit

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/structpb"
)

// GetAppEngineLog queries Google Cloud Logging entries for the given GCP
// project, returning each entry as a JSON string (newest first). The name
// is historical — logadmin reads any Cloud Logging source, not just App
// Engine.
func GetAppEngineLog(projectID string, filter string) ([]string, error) {
	ctx := context.Background()

	client, err := logadmin.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	it := client.Entries(ctx,
		logadmin.Filter(filter),
		logadmin.NewestFirst(),
	)

	var logs []string
	for {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		switch payload := entry.Payload.(type) {
		case *structpb.Struct:
			data := payload.AsMap()

			jsonStr, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}

			logs = append(logs, string(jsonStr))

		default:
			logs = append(logs, fmt.Sprintf("%v", entry.Payload))
		}
	}

	return logs, nil
}
