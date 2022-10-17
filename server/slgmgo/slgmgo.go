package slgmgo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDB *mongo.Database

func MongoSetUp() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017").SetTimeout(time.Second))
	if err != nil {
		panic(err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		panic(fmt.Sprintf("mongo ping fail %s", err.Error()))
	}
	db := client.Database("dbm")
	// 建立mongo句柄
	mongoDB = db
}

type SequenceCol struct {
	SeqName  string `bson:"seqName"`
	SeqValue int    `bson:"sqeValue"`
}

func getNextSequenceValue(sequenceName string) (int, error) {
	ctx := context.Background()
	filter := bson.M{"seqName": sequenceName}
	update := bson.M{"$inc": bson.M{"sqeValue": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true)
	var sc SequenceCol
	err := mongoDB.Collection("sequence").FindOneAndUpdate(ctx, filter, update, opts).Decode(&sc)
	if err != nil && err != mongo.ErrNoDocuments {
		return 0, err
	}
	return sc.SeqValue, nil
}

type Replay struct {
	Id   int   `bson:"id" json:"id"`
	Grid []int `bson:"grid" json:"grid"`
	Dice []int `bson:"dice" json:"dice"`
}

func InsertReply(grid, dice []int) (int, error) {
	incrId, err := getNextSequenceValue("replay")
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	document := Replay{Id: incrId, Grid: grid, Dice: dice}
	opts := options.InsertOne()
	_, err = mongoDB.Collection("replay").InsertOne(ctx, document, opts)
	if err != nil {
		return 0, err
	}
	return incrId, nil
}

func FindAndUpdateReply(incrId, randNum int) error {
	ctx := context.Background()
	filter := bson.M{"id": incrId}
	update := bson.M{"$push": bson.M{"dice": randNum}}
	opts := options.FindOneAndUpdate().SetUpsert(false)
	return mongoDB.Collection("replay").FindOneAndUpdate(ctx, filter, update, opts).Err()
}

func FindOneReply(incrId int) (Replay, error) {
	ctx := context.Background()
	filter := bson.M{"id": incrId}
	opts := options.FindOne()
	var replay Replay
	err := mongoDB.Collection("replay").FindOne(ctx, filter, opts).Decode(&replay)
	if err != nil {
		return Replay{}, err
	}
	return replay, nil
}
