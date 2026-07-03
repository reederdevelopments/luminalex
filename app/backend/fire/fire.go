package fire

import (
	"context"

	"cloud.google.com/go/firestore"
)

func Store(ctx context.Context, googleID string, dbID string) (*firestore.Client, error) {
	client, err := firestore.NewClientWithDatabase(ctx, googleID, dbID)
	if err != nil {
		return nil, err
	}
	return client, nil
}
