package main

import (
	"errors"
	"strings"

	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
)

const (
	TOPIC_LENGTH_LIMIT   = 36
	TITLE_LENGTH_LIMIT   = 100
	CONTENT_LENGTH_LIMIT = 3000
	USER_BUCKET          = "USER"
)

var (
	ErrContentLengthTooLong = errors.New("the length of topic or title or content is larger than limitation")
)

type content struct {
	UserId  string `json:"user_id" validte:"required"`
	Title   string `json:"title" validte:"required"`
	Topic   string `json:"topic" validte:"required"`
	Content string `json:"content" validte:"required"`
}

type shortContent struct {
}

func (sc *shortContent) Initialize(ctx code.Context) code.Response {
	return code.OK([]byte("ok"))
}

func (sc *shortContent) StoreShortContent(ctx code.Context) code.Response {
	args := content{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	userKey := USER_BUCKET + "/" + args.UserId + "/" + args.Topic + "/" + args.Title
	if len(args.Topic) > TOPIC_LENGTH_LIMIT ||
		len(args.Title) > TITLE_LENGTH_LIMIT ||
		len(args.Content) > CONTENT_LENGTH_LIMIT {
		return code.Error(ErrContentLengthTooLong)
	}
	if err := ctx.PutObject([]byte(userKey), []byte(args.Content)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok"))
}

func (sc *shortContent) QueryByUser(ctx code.Context) code.Response {
	args := struct {
		UserID string `json:"user_id" validte:"required"`
	}{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	prefix := USER_BUCKET + "/" + args.UserID + "/"
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()
	contents := []content{}
	for iter.Next() {
		value := strings.Split(string(iter.Key()[len(USER_BUCKET):]), "/")
		contents = append(contents, content{
			UserId:  args.UserID,
			Topic:   value[0],
			Title:   value[1],
			Content: string(iter.Value()),
		})
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(contents)
}

func (sc *shortContent) QueryByTitle(ctx code.Context) code.Response {
	args := struct {
		UserId string `json:"user_id" validte:"required"`
		Topic  string `json:"topic" validte:"required"`
		Title  string `json:"title" validte:"required"`
	}{}
	if err := utils.Unmarshal(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	value, err := ctx.GetObject([]byte(USER_BUCKET + "/" + args.UserId + "/" + args.Topic + "/" + args.Title))
	if err != nil {
		return code.Error(err)
	}
	return code.OK(value)
}

func (sc *shortContent) QueryByTopic(ctx code.Context) code.Response {
	args := struct {
		UserId string `json:"user_id" validte:"required"`
		Topic  string `json:"topic" validte:"required"`
	}{}
	prefix := USER_BUCKET + "/" + args.UserId + "/" + args.Topic + "/"
	iter := ctx.NewIterator(code.PrefixRange([]byte(prefix)))
	defer iter.Close()

	contents := []content{}

	for iter.Next() {
		contents = append(contents, content{
			UserId:  args.UserId,
			Topic:   args.Topic,
			Title:   string(iter.Key())[len(prefix):],
			Content: string(iter.Value()),
		})
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.JSON(contents)
}

func main() {
	driver.Serve(new(shortContent))
}
