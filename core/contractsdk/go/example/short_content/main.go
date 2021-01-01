package main

import (
	"errors"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/code"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/driver"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/utils"
	"strings"
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

type shortContent struct {
}

func (sc *shortContent) Initialize(ctx code.Context) code.Response {
	return code.OK([]byte("ok~"))
}

func (sc *shortContent) StoreShortContent(ctx code.Context) code.Response {
	args := struct {
		UserId  string `json:"user_id" required:"true"`
		Title   string `json:"title" required:"true"`
		Topic   string `json:"topic" required:"true"`
		Content string `json:"content" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	userKey := USER_BUCKET + "/" + args.UserId + "/" + args.Topic + "/" + args.Title
	if len(args.Topic) > TOPIC_LENGTH_LIMIT || len(args.Title) > TITLE_LENGTH_LIMIT ||
		len(args.Content) > CONTENT_LENGTH_LIMIT {
		return code.Error(ErrContentLengthTooLong)
	}
	if err := ctx.PutObject([]byte(userKey), []byte(args.Content)); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte("ok~"))
}

func (sc *shortContent) QueryByUser(ctx code.Context) code.Response {
	args := struct {
		UserID string `json:"user_id" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
		return code.Error(err)
	}
	start := USER_BUCKET + "/" + args.UserID + "/"
	end := start + "~"
	builder := strings.Builder{}
	iter := ctx.NewIterator([]byte(start), []byte(end))
	defer iter.Close()
	for iter.Next() {
		builder.Write(iter.Key())
		builder.WriteString("\n")
		builder.Write(iter.Value())
		builder.WriteString("\n")
	}
	if err := iter.Error(); err != nil {
		return code.Error(err)
	}
	return code.OK([]byte(builder.String()))
}
func (sc *shortContent) QueryByTitle(ctx code.Context) code.Response {
	args := struct {
		UserId string `json:"user_id" required:"true"`
		Topic  string `json:"topic" required:"true"`
		Title  string `json:"title" required:"true"`
	}{}
	if err := utils.Validate(ctx.Args(), &args); err != nil {
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
		UserId string `json:"user_id" required:"true"`
		Topic  string `json:"user_id" required:"true"`
	}{}
	start := USER_BUCKET + "/" + args.UserId + "/" + args.Topic + "/"
	end := start + "~"
	iter := ctx.NewIterator([]byte(start), []byte(end))
	defer iter.Close()

	builder := strings.Builder{}
	for iter.Next() {
		builder.Write(iter.Key())
		builder.WriteString("\n")
		builder.Write(iter.Value())
		builder.WriteString("\n")
	}
	return code.OK([]byte(builder.String()))
}

func main() {
	driver.Serve(new(shortContent))
}
