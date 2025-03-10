package dao

import (
	"context"
	"coolcar/shared/id"
	mgutil "coolcar/shared/mongo"
	"coolcar/shared/mongo/objid"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	col *mongo.Collection
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		col: db.Collection("blob"),
	}
}

type BlobRecord struct {
	mgutil.IDField `bson:"inline"`
	AccountID      string `bson:"accountid"`
	Path           string `bson:"path"`
}

// CreateBlob creates a blob.
func (m *Mongo) CreateBlob(c context.Context, aid id.AccountID) (*BlobRecord, error) {
	br := &BlobRecord{
		AccountID: aid.String(),
	}
	objID := mgutil.NewObjID()
	br.ID = objID
	// 每个account有一个目录，目录下有一个文件，文件名为objID.Hex()
	br.Path = fmt.Sprintf("%s/%s", aid.String(), objID.Hex())

	_, err := m.col.InsertOne(c, br)
	if err != nil {
		return nil, err
	}
	return br, nil
}

// GetBlob gets a blob by id.
func (m *Mongo) GetBlob(c context.Context, bid id.BlobID) (*BlobRecord, error) {
	objID, err := objid.FromID(bid)
	if err != nil {
		return nil, fmt.Errorf("invalid object id: %v", err)
	}
	res := m.col.FindOne(c, bson.M{
		mgutil.IDFieldName: objID,
	})
	if err := res.Err(); err != nil {
		return nil, err
	}

	var br BlobRecord
	err = res.Decode(&br)
	if err != nil {
		return nil, fmt.Errorf("cannot decode blob record: %v", err)
	}
	return &br, nil
}
